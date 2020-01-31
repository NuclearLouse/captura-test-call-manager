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

func (api *assureAPI) sysName(db *gorm.DB) string {
	db.Take(api)
	return api.SystemName
}

func (api assureAPI) checkAuth(db *gorm.DB) bool {
	res, err := api.requestGET(api.Version)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}
	if string(body) == "" {
		return false
	}
	return true
}

func (assureAPI) parseBNumbers(customBNumbers string) (nums []int64) {
	bnums := strings.Split(customBNumbers, "\n")
	for _, n := range bnums {
		p, _ := strconv.ParseInt(strings.TrimPrefix(n, "+"), 10, 64)
		nums = append(nums, p)
	}
	return
}

func (assureAPI) newTestBnumbers(ttn string, nit foundTest, nums []int64) testSetBnumbers {
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

func (assureAPI) newTestDestination(ttn string, nit foundTest) testSetDestination {
	return testSetDestination{
		NoOfExecutions: nit.TestCalls,
		TestSetItems: []batchDestination{batchDestination{
			TestTypeName:  ttn,
			RouteID:       nit.TestSysRouteID,
			DestinationID: nit.DestinationID},
		},
	}
}

func (api assureAPI) buildNewTests(ttn string, nit foundTest) (interface{}, error) {
	if nit.BNumber != "" {
		bnums := api.parseBNumbers(nit.BNumber)
		return api.newTestBnumbers(ttn, nit, bnums), nil
	}

	if nit.TestCalls == 0 {
		return struct{}{}, errors.New("zero calls initialized")
	}
	return api.newTestDestination(ttn, nit), nil
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

	newTest, err := api.buildNewTests(ttn, nit)
	if err != nil {
		return err
	}
	jsonBody, err := json.Marshal(newTest)
	if err != nil {
		return err
	}
	log.Debug("Build request body: ", string(jsonBody))
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
	log.Infof("Successful run test. TestID:%d", newTests.TestBatchID)
	return nil
}

func (api assureAPI) checkTestComplete(db *gorm.DB, lt foundTest) error {
	testid := lt.TestingSystemRequestID
	log.Debugf("Sending a request Complete_Test system %s for test_id %s", api.SystemName, testid)
	res, err := api.requestGET(api.StatusTests + testid)
	if err != nil {
		return err
	}
	log.Debugf("Successful response to the request Complete_Test for system %s test_ID %s", api.SystemName, testid)
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
		// 0 - Unknown
		// 1 - Created
		// 2 - Waiting
		// 3 - Running
		log.Debug("Wait. The test is not over yet for test_ID", testid)
		return nil
	case 4:
		// 4 - Finishing
		log.Debug("The end test for test_ID", testid)
		req := fmt.Sprintf("%sTest+Details+:+FAS+-+VQ+-+with+audio&Par1=%s", api.QueryResults, testid)
		res, err := api.requestGET(req)
		log.Debugf("Sending request TestResults for system %s test_ID %s", api.SystemName, testid)
		if err != nil {
			return err
		}
		log.Infof("Successful response to the request TestResults for system %s test_ID %s", api.SystemName, testid)
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		var callsinfo TestBatchResults
		if err := json.Unmarshal(body, &callsinfo); err != nil {
			return err
		}
		start := time.Now()
		log.Debugf("Start transaction insert into the table TestResults for system Assure test_id %s", testid)
		if err := api.insertCallsInfo(db, callsinfo, lt); err != nil {
			return err
		}
		log.Infof("Successfully insert data from table TestResults for system Assure test_ID %s", testid)
		log.Debug("Elapsed time insert transaction", time.Since(start))
		statistics = callsStatistics(db, testid)
		statistics.TestedFrom = assureParseTime(result.Created)
		statistics.TestedByUser = lt.RequestByUser
		statistics.TestResult = "OK"
		if err = db.Model(&statistics).Where(`"TestingSystemRequestID"=?`, testid).Update(statistics).Error; err != nil {
			return err
		}
		log.Info("Successfully update data to the table Purch_Oppt from test_ID", testid)
		go api.checkPresentAudioFile(db, callsinfo)
		return nil
	case 5, 6, 7:
		// 5 - Cancelling
		// 6 - Cancelled
		// 7 - Exception
		log.Info("Cancelled test for test_ID", testid)
		statistics.RequestState = 2
		statistics.TestedUntil = time.Now()
		statistics.TestComment = "Cancelled test by Assure for test_ID" + testid
		if err = db.Model(&statistics).Where(`"TestingSystemRequestID"=?`, testid).Update(statistics).Error; err != nil {
			return err
		}
		log.Debug("Successfully update data to the table Purch_Oppt from test_ID", testid)
		return nil
	}
	return nil
}

func (api assureAPI) checkPresentAudioFile(db *gorm.DB, tr TestBatchResults) {
	for _, res := range tr.QueryResult1 {
		callID := fmt.Sprintf("%d", res.CallResultID)

		if res.CallDuration == 0 || res.APartyAudio == "" {
			log.Info("Not present audio file for call_id", callID)
			if err := insertEmptyFiles(db, callID); err != nil {
				log.Errorf(603, "Cann't update data row about empty request for call_id %s|%v", callID, err)
			}
			continue
		}
		content, err := hex.DecodeString(res.APartyAudio)
		if err != nil {
			log.Errorf(605, "Error decode field APartyAudio for call_id %s|%v", callID, err)
			continue
		}

		if err := createGZ(content, callID); err != nil {
			log.Errorf(606, "Error create GZ file for call_id %s|%v", callID, err)
			continue
		}

		if err = uncompressGZ(callID); err != nil {
			log.Errorf(607, "Error uncompress GZ file for call_id %s|%v", callID, err)
			continue
		}
		var cWav []byte
		cWav, err = decodeToWAV(callID, "amr")
		if err != nil || len(cWav) == 0 {
			log.Errorf(608, "Error decode to wav file for call_id %s|%v", callID, err)
			cWav = []byte("C&V:Cann't decode to wav file")
			// тут нужна проверка на очистку временной папки и вставка этой записи в таблицу
			continue
		}
		log.Info("Created WAV file for call_id", callID)

		x, _ := strconv.Atoi(fmt.Sprintf("%.f", 500*res.BConnectTime/res.BDisconnectTime))

		cImg, err := waveFormImage(callID, x)
		if err != nil || len(cImg) == 0 {
			log.Errorf(609, "Cann't create waveform image file for call_id %s|%v", callID, err)
			cImg = labelEmptyBMP("C&V:Cann't create waveform image file")
			// тут нужна проверка на очистку временной папки и вставка этой записи в таблицу
			continue
		}
		log.Info("Created image PNG file for call_id", callID)

		listDeleteFiles := []string{
			srvTmpFolder + callID + ".amr.gz",
			srvTmpFolder + callID + ".amr",
			srvTmpFolder + callID + ".wav",
			srvTmpFolder + callID + ".png",
			srvTmpFolder + callID + ".bmp",
		}

		callsinfo := CallingSysTestResults{
			DataLoaded: true,
			AudioFile:  cWav,
			AudioGraph: cImg,
		}
		if err = updateCallsInfo(db, callID, callsinfo); err != nil {
			log.Errorf(610, "Cann't insert WAV file into table for system Assure call_id %s|%v", callID, err)
			continue
		}
		log.Info("Insert WAV and IMG file for callid", callID)

		if err = deleteFiles(listDeleteFiles); err != nil {
			log.Errorf(611, "Cann't delete some files for call_id %s|%v", callID, err)
		}
	}
}

func (assureAPI) insertCallsInfo(db *gorm.DB, tr TestBatchResults, lt foundTest) error {
	for _, res := range tr.QueryResult1 {
		callstart := assureParseTime(res.TestStartTime)
		callinfo := CallingSysTestResults{
			AudioURL:                 strconv.Itoa(res.CallResultID),
			CallID:                   strconv.Itoa(res.CallResultID),
			CallListID:               lt.TestingSystemRequestID,
			TestSystem:               lt.SystemID,
			CallType:                 res.TestType, // or string(lt.TestType),
			Destination:              res.BNetwork,
			CallStart:                callstart,
			CallComplete:             time.Unix(callstart.Unix()+int64(res.DisconnectTime), 0),
			CallDuration:             res.CallDuration,
			AlertTime:                res.BAlertTime,
			ConnectTime:              res.BConnectTime,
			BNumber:                  res.Bnumber,
			Route:                    res.Route, // or lt.RouteCarrier
			Status:                   res.ReleaseCause,
			CliDetectedCallingNumber: res.CLIDelivered,
			CliResult:                res.Result,
			VoiceQualityMos:          res.MOSA,
			PDD:                      res.PGRD,
			CallingNumber:            res.ANumber,
		}
		if err := db.Create(&callinfo).Error; err != nil {
			return err
		}
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
