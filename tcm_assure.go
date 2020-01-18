// assure.go
//
// The file contains the functions necessary for the operation of the "Assure" system.
//
// Must be present functions that satisfy the tester interface
//
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
	log "redits.oculeus.com/asorokin/my_packages/logging"
)

type testSetDestination struct {
	NoOfExecutions int
	TestSetItems
}

type TestSetItems struct {
	TestTypeName  string
	RouteID       int
	DestinationID int
}

type testSetBnumbers struct {
	TestSetItems []batchBnumbers
}

type batchBnumbers struct {
	TestTypeName string
	RouteID      int
	PhoneNumber  int64
}

func newTestBnumbers(ttn string, nit foundTest, nums []int64) testSetBnumbers {
	var batches []batchBnumbers
	for _, p := range nums {
		b := batchBnumbers{
			TestTypeName: ttn,
			RouteID:      nit.TestSysRouteID,
			PhoneNumber:  p}

		batches = append(batches, b)
	}
	return testSetBnumbers{
		TestSetItems: batches,
	}
}

func newTestDestination(ttn string, nit foundTest) testSetDestination {
	return testSetDestination{
		NoOfExecutions: nit.TestCalls,
		TestSetItems: TestSetItems{
			TestTypeName:  ttn,
			RouteID:       nit.TestSysRouteID,
			DestinationID: nit.DestinationID},
	}
}

func buildNewTests(ttn string, nit foundTest) (interface{}, error) {
	if nit.BNumber != "" {
		bnums := parseBNumbers(nit.BNumber)
		return newTestBnumbers(ttn, nit, bnums), nil
	}

	if nit.TestCalls == 0 {
		return struct{}{}, errors.New("zero calls initialized")
	}
	return newTestDestination(ttn, nit), nil
}

func (api assureAPI) requestGET(r string) (*http.Response, error) {
	req, err := http.NewRequest("GET", api.URL+r, nil)
	if err != nil {
		return nil, err
	}
	res, err := api.httpRequest(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api assureAPI) requestPOST(r string, jsonStr []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", api.URL+r, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "text/json")
	if err != nil {
		return nil, err
	}
	res, err := api.httpRequest(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api assureAPI) httpRequest(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(api.User, api.Pass)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api *assureAPI) sysName(db *gorm.DB) string {
	db.Take(api)
	return api.SystemName
}

func (api assureAPI) prepareRequests(db *gorm.DB, interval int64) {
	for {
		log.Info("Send preparatory requests for", api.SystemName)
		log.Debug("API Settings", api)
		httpRequests := map[string]string{
			"assure_routes":             api.Routes,
			"assure_destinations":       api.Destinations,
			"assure_nodes":              api.Nodes,
			"assure_nodes_capabilities": api.NodesCapabilities,
		}
		keys := make([]string, 0)
		for key := range httpRequests {
			keys = append(keys, key)
		}
		for i := range keys {
			var err error
			res, err := api.requestGET(api.QueryResults + httpRequests[keys[i]])
			// log.Debug("Prepare response", res)
			if err != nil {
				log.Errorf(600, "Failed to get a response to the request %s|%v", keys[i], err)
				continue
			}
			log.Info("Successful response to the request", keys[i])
			start := time.Now()
			log.Debug("Start transaction insert into the table", keys[i])
			if err := insertsPrepareJSON(db, keys[i], res); err != nil {
				log.Errorf(601, "Could not insert data from response %s|%v", keys[i], err)
				continue
			}
			log.Info("Successfully insert data from response", keys[i])
			log.Debugf("Elapsed time transaction insert %s %v", keys[i], time.Since(start))
		}
		log.Infof("The next data update to prepare %s after %d hours", api.SystemName, interval)
		time.Sleep(time.Duration(interval) * time.Hour)
	}
}

func insertsPrepareJSON(db *gorm.DB, req string, res *http.Response) error {
	body, _ := ioutil.ReadAll(res.Body)
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return err
	}
	var bulkslice []interface{}
	switch req {
	case "assure_routes":
		var route assureRoutes // table's struct
		var routes Routes      // JSON's struct
		if err := db.Delete(assureRoutes{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", route.TableName())
		if err := json.Unmarshal(body, &routes); err != nil {
			return err
		}
		for i := range routes.QueryResult1 {
			if err := mapstructure.Decode(routes.QueryResult1[i], &route); err != nil {
				return err
			}
			if dialectDB == "sqlite3" {
				if err := tx.Create(&route).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, route)
		}
	case "assure_destinations":
		var destination assureDestinations
		var destinations Destinations
		if err := db.Delete(assureDestinations{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", destination.TableName())
		if err := json.Unmarshal(body, &destinations); err != nil {
			return err
		}
		for i := range destinations.QueryResult1 {
			if err := mapstructure.Decode(destinations.QueryResult1[i], &destination); err != nil {
				return err
			}
			if dialectDB == "sqlite3" {
				if err := tx.Create(&destination).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, destination)
		}
	case "assure_nodes":
		var node assureNodes
		var nodes Nodes
		if err := db.Delete(assureNodes{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", node.TableName())
		if err := json.Unmarshal(body, &nodes); err != nil {
			return err
		}
		for i := range nodes.QueryResult1 {
			if err := mapstructure.Decode(nodes.QueryResult1[i], &node); err != nil {
				return err
			}
			if dialectDB == "sqlite3" {
				if err := tx.Create(&node).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, node)
		}
	case "assure_nodes_capabilities":
		var node assureNodesCapabilities
		var nodes NodeCapabilities
		if err := db.Delete(assureNodesCapabilities{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", node.TableName())
		if err := json.Unmarshal(body, &nodes); err != nil {
			return err
		}
		for i := range nodes.QueryResult1 {
			if err := mapstructure.Decode(nodes.QueryResult1[i], &node); err != nil {
				return err
			}
			if dialectDB == "sqlite3" {
				if err := tx.Create(&node).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, node)
		}
	}
	switch dialectDB {
	case "sqlite3":
		err := tx.Commit().Error
		if err != nil {
			return err
		}
	default:
		if err := gormbulk.BulkInsert(db, bulkslice, 1500); err != nil {
			return err
		}
	}
	return nil
}

func (api assureAPI) runNewTest(db *gorm.DB, nit foundTest) error {
	var ttn string

	switch nit.TestType.name() {
	case "cli":
		ttn = "CLI"
	case "voice":
		ttn = "Voice Quality Basic"
	case "fas":
		ttn = "FAS"
	}

	newTest, err := buildNewTests(ttn, nit)
	if err != nil {
		return err
	}
	jsonBody, err := json.Marshal(newTest)
	if err != nil {
		return err
	}
	response, err := api.requestPOST(api.StatusTests, jsonBody)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var newTests TestBatches
	if err := json.Unmarshal(body, &newTests); err != nil {
		return err
	}
	response.Body.Close()

	log.Debug(string(body))

	if newTests.TestBatchID == 0 {
		err := errors.New("no return TestingSystemRequestID")
		testinfo := PurchOppt{TestingSystemRequestID: "0"}
		testinfo.failedTest(db, nit.RequestID, string(body))
		return err
	}

	newTestInfo := PurchOppt{
		TestingSystemRequestID: strconv.Itoa(newTests.TestBatchID),
		RequestState:           2}
	if err := db.Model(&newTestInfo).Where(`"RequestID"=?`, nit.RequestID).Update(newTestInfo).Error; err != nil {
		return err
	}

	return nil
}

func (api assureAPI) uploadResultFiles(db *gorm.DB) {
	for {
		sysname := api.SystemName
		var rows []CallingSysTestResults
		if err := db.Where(`"DataLoaded"=false AND "TestSystem"=?`, api.SystemID).Find(&rows).Error; err != nil {
			log.Errorf(602, "Cann't obtain rows for DataLoaded=false|%v", err)
			continue
		}
		if len(rows) == 0 {
			log.Trace("Not rows for files_uploaded=false. All files upload.")
			continue
		}
		testInProgress := " "
		for i := range rows {
			if rows[i].CallDuration == 0 {
				log.Info("Not present audio file for call_id", rows[i].CallID)
				if err := insertEmptyFiles(db, rows[i].CallID); err != nil {
					log.Errorf(603, "Cann't update data row about empty request for call_id %s|%v", rows[i].CallID, err)
				}
				continue
			}
			if testInProgress == rows[i].CallListID {
				continue
			}
			err := api.checkPresentAudioFile(db, rows[i])
			if err != nil {
				log.Errorf(604, "Error for check present audio file for system %s and call_id %s|%v", sysname, rows[i].CallID, err)
				continue
			}
			testInProgress = rows[i].CallListID
		}
		log.Infof("All present files download from %s server and upload into the table TestResults", sysname)
		time.Sleep(time.Duration(1) * time.Hour)
	}
}

func (api assureAPI) checkTestComplete(db *gorm.DB, lt foundTest) error {
	sysname := api.SystemName
	testid := lt.TestingSystemRequestID
	log.Debugf("Sending a request Complete_Test system %s for test_id %s", sysname, testid)
	res, err := api.requestGET(api.StatusTests + testid)
	if err != nil {
		return err
	}
	log.Debugf("Successful response to the request Complete_Test for system %s test_ID %s", sysname, testid)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()
	var result TestBatches
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	var statistics PurchOppt
	switch result.StatusID {
	case 0, 1, 2, 3:
		return nil
	case 4:
		res, err := api.requestGET(api.TestResults + testid)
		log.Debugf("Sending request TestResults fot system %s test_ID %s", sysname, testid)
		if err != nil {
			return err
		}
		log.Infof("Successful response to the request TestResults for system %s test_ID %s", sysname, testid)
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		res.Body.Close()
		var callsinfo TestBatchResults
		if err := json.Unmarshal(body, &callsinfo); err != nil {
			return err
		}
		start := time.Now()
		log.Debugf("Start transaction insert into the table TestResults for system %s test_id %s", sysname, testid)
		if err := api.insertCallsInfo(db, callsinfo, lt); err != nil {
			return err
		}
		log.Infof("Successfully insert data from table TestResults for system %s test_ID %s", sysname, testid)
		log.Debug("Elapsed time insert transaction", time.Since(start))
		statistics = callsStatistics(db, testid)
		statistics.TestedFrom = result.ParseTime(result.Created)
		statistics.TestedByUser = lt.RequestByUser
		if err = db.Model(&statistics).Where(`"TestingSystemRequestID"=?`, testid).Update(statistics).Error; err != nil {
			return err
		}
		log.Info("Successfully update data to the table Purch_Oppt from test_ID", testid)
		return nil
	case 5, 6, 7:
		log.Info("Cancelled test for test_ID", testid)
		statistics.RequestState = 2
		statistics.TestedUntil = time.Now()
		statistics.TestComment = "Cancelled test by Assure for test_ID" + testid
		if err = db.Model(&statistics).Where(`"TestingSystemRequestID"=?`, testid).Update(statistics).Error; err != nil {
			return err
		}
		log.Info("Successfully update data to the table Purch_Oppt from test_ID", testid)
		return nil
	}
	log.Info("Wait. The test is not over yet for test_ID", testid)
	return nil
}

func (api assureAPI) checkPresentAudioFile(db *gorm.DB, ctr CallingSysTestResults) error {
	req := fmt.Sprintf("%sTest+Details+:+FAS+-+VQ+-+with+audio&Par1=%s", api.QueryResults, ctr.CallListID)
	res, err := api.requestGET(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil && err != io.EOF {
		return err
	}
	var result queryResults
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	var name string
	for i := range result.QueryResult1 {
		partAudio := result.QueryResult1[i].APartyAudio
		if partAudio != "" {
			content, err := hex.DecodeString(partAudio)
			if err != nil {
				log.Errorf(605, "Error decode field APartyAudio for call_id %d|%v", result.QueryResult1[i].CallResultID, err)
				continue
			}

			name = fmt.Sprintf("%d", result.QueryResult1[i].CallResultID)
			if err := createGZ(content, name); err != nil {
				log.Errorf(606, "Error create GZ file for call_id %s|%v", name, err)
				continue
			}

			if err = uncompressGZ(name); err != nil {
				log.Errorf(607, "Error uncompress GZ file for call_id %s|%v", name, err)
				continue
			}
			var cWav []byte
			cWav, err = decodeToWAV(name, "amr")
			if err != nil || len(cWav) == 0 {
				log.Errorf(608, "Error decode to wav file for call_id %s|%v", name, err)
				cWav = []byte("C&V:Cann't decode to wav file")
				// тут нужна проверка на очистку временной папки и вставка этой записи в таблицу
				continue
			}
			log.Info("Created WAV file for call_id", name)

			cImg, err := waveFormImage(name, 0)
			if err != nil || len(cImg) == 0 {
				log.Errorf(609, "Cann't create waveform image file for call_id %s|%v", name, err)
				cImg = []byte("C&V:Cann't create waveform image file")
				// тут нужна проверка на очистку временной папки и вставка этой записи в таблицу
				continue
			}
			log.Info("Created image PNG file for call_id", name)

			listDeleteFiles := []string{
				srvTmpFolder + name + ".amr.gz",
				srvTmpFolder + name + ".amr",
				srvTmpFolder + name + ".wav",
				srvTmpFolder + name + ".png",
				srvTmpFolder + name + ".bmp",
			}

			callsinfo := CallingSysTestResults{
				DataLoaded:  true,
				AudioFile:   cWav,
				AudioGraph:  cImg,
				ConnectTime: result.QueryResult1[i].PGAD,
				CallType:    result.QueryResult1[i].TestType,
			}
			if err = updateCallsInfo(db, name, callsinfo); err != nil {
				log.Errorf(610, "Cann't insert WAV file into table for system %s call_id %s|%v", api.SystemName, name, err)
				continue
			}
			log.Info("Insert WAV and IMG file for callid", name)

			if err = deleteFiles(listDeleteFiles); err != nil {
				log.Errorf(611, "Cann't delete some files for call_id %s|%v", name, err)
			}
		}
	}
	return nil
}

func (assureAPI) insertCallsInfo(db *gorm.DB, tr TestBatchResults, lt foundTest) error {
	var err error
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return err
	}
	for i := range tr.TestBatchResult1 {
		res := tr.TestBatchResult1[i]
		callstart := tr.ParseTime(res.TestStartTime)
		// тут надо продумать получение полей с комментариями по Result, CLI и т.д.
		r := res.Result
		if len(r) > 30 {
			r = r[:30]
		}
		callinfo := CallingSysTestResults{
			AudioURL:                 strconv.Itoa(res.CallResultID),
			CallID:                   strconv.Itoa(res.CallResultID),
			CallListID:               lt.TestingSystemRequestID,
			TestSystem:               lt.SystemID,
			CallType:                 string(lt.TestType),
			Destination:              res.BNetwork,
			CallStart:                callstart,
			CallComplete:             time.Unix(callstart.Unix()+int64(res.DisconnectTime), 0),
			CallDuration:             res.CallDuration,
			AlertTime:                res.BAlertTime,
			ConnectTime:              res.BConnectTime,
			BNumber:                  res.BTestNode,
			Route:                    res.Route,
			Status:                   res.ReleaseCause,
			CliDetectedCallingNumber: res.CLIDelivered,
			CliResult:                r,
			VoiceQualityMos:          res.MOSA,
			VoiceQualitySNR:          int(res.SNR),
			VoiceQualitySpeechLevel:  int(res.SpeechLevel),
		}
		if err := tx.Create(&callinfo).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit().Error
	if err != nil {
		return err
	}
	return nil
}

func createGZ(content []byte, nameFile string) error {
	fileGZ := srvTmpFolder + nameFile + ".amr.gz"
	file, err := os.Create(fileGZ)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	if err != nil {
		return err
	}
	return nil
}

func uncompressGZ(name string) error {
	pathGZ := srvTmpFolder + name + ".amr.gz"
	gzipFile, err := os.Open(pathGZ)
	if err != nil {
		return err
	}
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	unzipNameFile := strings.Split(pathGZ, ".gz")
	outfileWriter, err := os.Create(unzipNameFile[0])
	if err != nil {
		return err
	}
	io.Copy(outfileWriter, gzipReader)
	outfileWriter.Close()
	gzipReader.Close()
	gzipFile.Close()

	return nil
}

// GET requests for init new test
// request = fmt.Sprintf("%sTestTypeName=%s&RouteID=%d&PhoneNumber=%s",
// 	api.NewTestGet,
// 	ttn,
// 	nit.TestSysRouteID,
// 	strings.TrimPrefix(nit.BNumber, "+"))
// request = fmt.Sprintf("%sTestTypeName=%s&RouteID=%d&DestinationID=%d&NoOfExecutions=%d",
// 	api.NewTestGet,
// 	ttn,
// 	nit.TestSysRouteID,
// 	nit.DestinationID,
// 	nit.TestCalls)
