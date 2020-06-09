// tcm_assure.go
//
// The file contains the functions necessary for the operation of the "Assure" system.
//
// Must be present functions that satisfy the tester interface
//
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "captura_tcm/logger"

	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"

	gormbulk "github.com/t-tiger/gorm-bulk-insert"
)

func (api *assureAPI) sysName(db *gorm.DB) string {
	db.Take(api)
	return api.SystemName
}

func (api *assureAPI) sysID(db *gorm.DB) int {
	db.Take(api)
	return api.SystemID
}

func (api assureAPI) checkAuth(db *gorm.DB) bool {
	res, err := api.newRequest("GET", api.Version, nil)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	if res.Status == "401 Unauthorized" {
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

func (assureAPI) newTestBnumbers(nit foundTest, nums []int64) testSetBnumbers {
	var batches []batchBnumbers
	for _, p := range nums {
		b := batchBnumbers{
			TestTypeName: nit.TestType,
			RouteID:      nit.TestSysRouteID,
			PhoneNumber:  p,
		}

		batches = append(batches, b)
	}
	return testSetBnumbers{
		TestSetItems: batches,
	}
}

func (assureAPI) newTestDestination(nit foundTest) testSetDestination {
	return testSetDestination{
		NoOfExecutions: nit.TestCalls,
		TestSetItems: []batchDestination{
			{
				TestTypeName:  nit.TestType,
				RouteID:       nit.TestSysRouteID,
				DestinationID: nit.DestinationID,
			},
		},
	}
}

func (assureAPI) newTestSMS(nit foundTest) testSetSMS {
	return testSetSMS{
		NoOfExecutions: nit.TestCalls,
		TestSetItems: []batchSMS{
			{
				SMSRouteID:      nit.TestSysRouteID, // нужен route_id для SMS
				DestinationID:   nit.DestinationID,  // нужен destination_id для SMS
				SMSTemplateName: nit.SMSTemplate,
			},
		},
	}
}

func (api assureAPI) buildNewTests(nit foundTest) (interface{}, error) {
	if nit.BNumber != "" {
		bnums := api.parseBNumbers(nit.BNumber)
		return api.newTestBnumbers(nit, bnums), nil
	}

	if nit.TestCalls == 0 {
		return struct{}{}, errors.New("zero calls initialized")
	}

	if nit.TestType == "SMS" {
		return api.newTestSMS(nit), nil
	}

	return api.newTestDestination(nit), nil
}

func (api assureAPI) newRequest(method, request string, body []byte) (*http.Response, error) {
	log.Debugf("Request %s:%s", method, api.URL+request)
	req, err := http.NewRequest(method, api.URL+request, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if method == "POST" {
		log.Debug("Request Body:", string(body))
		req.Header.Set("Content-Type", "text/json")
	}
	req.SetBasicAuth(api.User, api.Pass)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api assureAPI) runNewTest(db *gorm.DB, nit foundTest) error {
	newTest, err := api.buildNewTests(nit)
	if err != nil {
		return err
	}
	jsonBody, err := json.Marshal(newTest)
	if err != nil {
		return err
	}
	log.Debug("Build request body: ", string(jsonBody))
	res, err := api.newRequest("POST", api.StatusTests, jsonBody)
	if err != nil {
		return err
	}

	var newTests testStatusAssure
	if err := json.NewDecoder(res.Body).Decode(&newTests); err != nil {
		return err
	}
	res.Body.Close()

	if newTests.TestBatchID == 0 {
		return errors.New("response not return TestingSystemRequestID")
	}

	testinfo := purchOppt{
		TestResult:             newTests.String(),
		TestingSystemRequestID: strconv.Itoa(newTests.TestBatchID),
		RequestState:           2}
	if err := testinfo.updateTestInfo(db, nit.RequestID); err != nil {
		return err
	}
	log.Infof("Successful run test. TestID:%d", newTests.TestBatchID)
	return nil
}

func (api assureAPI) checkTestComplete(db *gorm.DB, lt foundTest) error {
	testid := lt.TestingSystemRequestID
	log.Debugf("Sending a request Complete_Test system %s for test_id %s", api.SystemName, testid)
	res, err := api.newRequest("GET", api.StatusTests+testid, nil)
	if err != nil {
		return err
	}
	log.Debugf("Successful response to the request Complete_Test for system %s test_ID %s", api.SystemName, testid)

	var result testStatusAssure
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return err
	}
	res.Body.Close()

	var ti purchOppt
	ti.TestResult = result.String()

	switch result.StatusID {
	case 1, 2, 3: // Created, Waiting, Running
		log.Trace("Wait. The test is not over yet for test_ID", testid)
	case 4: // Finishing
		log.Info("The end test for test_ID", testid)

		var req string
		smstest := strings.Contains(strings.ToLower(lt.TestType), "sms")
		if smstest {
			// Возможны два варианта запросов для получения результатов SMS тестов:
			//**************** 1 ********************
			// По ID теста
			// https://h-54-246-182-248.csg-assure.com/api/TestBatchResults/61825
			req = fmt.Sprintf("%s%s%d", api.URL, api.TestResults, result.TestBatchID)

			//**************** 2 ********************
			// По дате
			// https://h-54-246-182-248.csg-assure.com/api/QueryResults2?code=Details+:+SMS+MT&From=2020-02-24&To=2020-02-24
			// req = fmt.Sprintf("%sDetails+:+SMS+MT&From=%s&To=%[2]s", api.QueryResults, time.Now().Format("2006-01-02"))
		} else {
			req = fmt.Sprintf("%sTest+Details+:+CLI+-+FAS+-+VQ+-+with+audio&Par1=%s", api.QueryResults, testid)
		}

		res, err := api.newRequest("GET", req, nil)
		log.Debugf("Sending request TestResults for system %s test_ID %s", api.SystemName, testid)
		if err != nil {
			return err
		}
		log.Infof("Successful response to the request TestResults for system %s test_ID %s", api.SystemName, testid)

		switch {
		case smstest:
			var smsinfo testResultAssureSMS
			if err := json.NewDecoder(res.Body).Decode(&smsinfo); err != nil {
				return err
			}
			res.Body.Close()

			start := time.Now()
			log.Debugf("Start transaction insert into the table SMSTestResults for system Assure test_id %s", testid)
			if err := api.insertSMSInfo(db, smsinfo, testid); err != nil {
				return err
			}
			log.Infof("Successfully insert data from table SMSTestResults for system Assure test_ID %s", testid)
			log.Debug("Elapsed time insert transaction", time.Since(start))

			ti.smsStatisticsAssure(db, testid)
			ti.TestedFrom = assureParseTime(result.Created, ".")
			ti.TestedByUser = lt.RequestByUser

		default:
			var callsinfo testResultAssure
			if err := json.NewDecoder(res.Body).Decode(&callsinfo); err != nil {
				return err
			}
			res.Body.Close()

			start := time.Now()
			log.Debugf("Start transaction insert into the table TestResults for system Assure test_id %s", testid)
			if err := api.insertCallsInfo(db, callsinfo, lt); err != nil {
				return err
			}
			log.Infof("Successfully insert data from table TestResults for system Assure test_ID %s", testid)
			log.Debug("Elapsed time insert transaction", time.Since(start))

			ti.callsStatistic(db, testid)
			ti.TestedFrom = assureParseTime(result.Created, ".")
			ti.TestedByUser = lt.RequestByUser
			go api.downloadAudioFiles(db, callsinfo)
		}

	case 0, 5, 6, 7: // Unknown, Cancelling, Cancelled, Exception
		log.Info("Cancelled test for test_ID", testid)
		ti.RequestState = 2
		ti.TestedUntil = time.Now()
		ti.TestComment = "The test ended with status:" + result.String()
	}

	if err := ti.updateStatistic(db, testid); err != nil {
		return err
	}
	log.Debug("Successfully update data to the table Purch_Oppt from test_ID", testid)
	return nil
}

func (api assureAPI) cancelTest(db *gorm.DB, testid string) error {
	log.Debugf("Sending a request Cancel Test for system %s and test_id %s", api.SystemName, testid)
	_, err := api.newRequest("DELETE", api.StatusTests+testid, nil)
	if err != nil {
		return err
	}

	return nil
}

func (api assureAPI) downloadAudioFiles(db *gorm.DB, tr testResultAssure) {
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
		var cWav, cMP3 []byte
		cMP3, err = decodeAudio(callID, "amr", "mp3", "source")
		if err != nil || len(cMP3) == 0 {
			log.Errorf(608, "Error decode to mp3 file for call_id %s|%v", callID, err)
			cMP3 = []byte("C&V:Cann't decode to mp3 file")
			continue
		}
		log.Info("Created MP3 file for call_id", callID)

		cWav, err = decodeAudio(callID, "mp3", "wav", "source")
		if err != nil || len(cWav) == 0 {
			log.Errorf(608, "Error decode to wav file for call_id %s|%v", callID, err)
			cWav = []byte("C&V:Cann't decode to wav file")
			continue
		}
		log.Info("Created WAV file for call_id", callID)

		var x int
		if res.TestType != "CLI" {
			// TODO: Формулу потом надо пересмотреть
			x, _ = strconv.Atoi(fmt.Sprintf("%.f", 500*res.BConnectTime/res.DisconnectTime))
		}
		pngImg, bmpImg, err := waveFormImage(callID, x)
		if err != nil || len(bmpImg) == 0 || len(pngImg) == 0 {
			log.Errorf(609, "Cann't create waveform image files for call_id %s|%v", callID, err)
			bmpImg = labelEmptyBMP("C&V:Cann't create waveform image file")
			continue
		}
		log.Info("Created image files for call_id", callID)

		callsinfo := callingSysTestResults{
			DataLoaded: true,
			AudioFile:  cWav,
			AudioGraph: bmpImg}
		if err = callsinfo.updateCallsInfo(db, callID); err != nil {
			log.Errorf(610, "Cann't insert audio and image file into TestResults table for call_id %s|%v", callID, err)
			continue
		}

		webinfo := testFilesWEB{
			Callid:     callID,
			Testsystem: api.SystemID,
			Diagram:    pngImg,
			Audiofile:  cMP3}
		if err = webinfo.insertWebInfo(db); err != nil {
			log.Errorf(611, "Cann't insert audio and image file into testfiles_web table for call_id %s|%v", callID, err)
			continue
		}
		log.Info("Insert audio and image file for callid", callID)
	}
}

func (assureAPI) insertCallsInfo(db *gorm.DB, tr testResultAssure, lt foundTest) error {
	for _, res := range tr.QueryResult1 {
		callstart := assureParseTime(res.TestStartTime, ".")
		callinfo := callingSysTestResults{
			AudioURL:                 strconv.Itoa(res.CallResultID),
			CallID:                   strconv.Itoa(res.CallResultID),
			CallListID:               lt.TestingSystemRequestID,
			TestSystem:               lt.SystemID,
			CallType:                 res.TestType,
			Destination:              res.BNetwork,
			CallStart:                callstart,
			CallComplete:             time.Unix(callstart.Unix()+int64(res.DisconnectTime), 0),
			CallDuration:             res.CallDuration,
			AlertTime:                res.BAlertTime,
			ConnectTime:              res.BConnectTime,
			BNumber:                  res.Bnumber,
			Route:                    res.Route,
			Status:                   res.ReleaseCause,
			CliDetectedCallingNumber: res.CLIDelivered,
			CliResult:                res.Result,
			VoiceQualityMos:          res.MOSA,
			PDD:                      res.PGRD,
			CallingNumber:            res.ANumber,
			FasResult:                res.Result,
		}
		if err := db.Create(&callinfo).Error; err != nil {
			return err
		}
	}
	return nil
}

func (assureAPI) insertSMSInfo(db *gorm.DB, tr testResultAssureSMS, testID string) error {
	for _, re := range tr.TestBatchResult {
		smsinfo := assureSMSResult{
			TestBatchID:               testID,
			CallBatchItemID:           re.CallBatchItemID,
			CallBatchStatusTypeID:     re.CallBatchStatusTypeID,
			CallBatchExceptionMessage: re.CallBatchExceptionMessage,
			UITestStatusID:            re.UITestStatusID,
			UITestStatusDisplay:       re.UITestStatusDisplay,
			Modified:                  assureParseTime(re.Modified, "."),
			HasAudio:                  re.HasAudio,
			HasSecondAudio:            re.HasSecondAudio,
			TestScenarioRuntimeUID:    re.TestScenarioRuntimeUID,
			SMSResultID:               re.SMSResultID,
			Result:                    re.Result,
			Network:                   re.Network,
			TestNode:                  re.TestNode,
			Supplier:                  re.Supplier,
			Route:                     re.Route,
			SentTime:                  assureParseTime(re.SentTime, "."),
			DelTimeOK:                 re.DelTimeOK,
			ContentOK:                 re.ContentOK,
			OAOK:                      re.OAOK,
			STONOK:                    re.STONOK,
			AlphaOK:                   re.AlphaOK,
			DelRepOK:                  re.DelRepOK,
			DelTimeSec:                re.DelTimeSec,
			SMSC:                      re.SMSC,
			SMSCOwner:                 re.SMSCOwner,
			ErrorMsg:                  re.ErrorMsg,
			OAS:                       re.OAS,
			OAR:                       re.OAR,
			STONS:                     re.STONS,
			STONR:                     re.STONR,
			AlphaS:                    re.AlphaS,
			AlphaR:                    re.AlphaR,
			SNumPlanS:                 re.SNumPlanS,
			SNumPlanR:                 re.SNumPlanR,
			TextS:                     re.TextS,
			TextR:                     re.TextR,
			NoOfSegS:                  re.NoOfSegS,
			NoOfSegR:                  re.NoOfSegR,
			RecTime:                   assureParseTime(re.RecTime, "."),
			SMSCRecTime:               assureParseTime(re.SMSCRecTime, "+"),
			DelRepTime:                assureParseTime(re.DelRepTime, "."),
			MessageID:                 re.MessageID,
			Template:                  re.Template,
			SMS:                       re.SMS,
			DeliveryDetails:           re.DeliveryDetails,
			DelTimeLimit:              assureParseTime(re.DelTimeLimit, "."),
			Duplicates:                re.Duplicates,
			APIResponse:               re.APIResponse,
			ResultRecTime:             assureParseTime(re.ResultRecTime, "."),
			UDHS:                      re.UDHS,
			UDHR:                      re.UDHR,
			PDUR:                      re.PDUR,
			ResultTrace:               re.ResultTrace,
			ExceptionMsg:              re.ExceptionMsg,
			CustomerID:                re.CustomerID,
			SMSID:                     re.SMSID,
			SMSIDExpireTime:           assureParseTime(re.SMSIDExpireTime, "."),
			SMSTemplateID:             re.SMSTemplateID,
			SMSTemplateModified:       re.SMSTemplateModified,
			SubmitSMSResponseDelay:    re.SubmitSMSResponseDelay,
			DeliveryUpdateDelay:       re.DeliveryUpdateDelay,
			SMSReceiveDelay:           re.SMSReceiveDelay,
			SMSResultReceiveID:        re.SMSResultReceiveID,
			SMSResultIDR:              re.SMSResultIDR,
			PoPIDR:                    re.PoPIDR,
			AdapterInstanceIDR:        re.AdapterInstanceIDR,
			AdapterInstanceNameR:      re.AdapterInstanceNameR,
			DCSCharacterSetR:          re.DCSCharacterSetR,
			SMR:                       re.SMR,
			SegmentDuplicate:          re.SegmentDuplicate,
			SMSResultSendID:           re.SMSResultSendID,
			SMSResultIDS:              re.SMSResultIDS,
			PoPIDS:                    re.PoPIDS,
			AdapterInstanceIDS:        re.AdapterInstanceIDS,
			SubmitSMSResponseTime:     assureParseTime(re.SubmitSMSResponseTime, "."),
			StatusUpdateCode:          re.StatusUpdateCode,
			StatusUpdateTime:          assureParseTime(re.StatusUpdateTime, "."),
		}
		if err := db.Create(&smsinfo).Error; err != nil {
			return err
		}
	}
	return nil
}

func (api assureAPI) prepareRequests(db *gorm.DB, interval int64) {
	for {
		log.Info("Send preparatory requests for", api.SystemName)
		log.Debug("API Settings", api)
		// TODO: надо добавить таблицы нужные для SMS тестов
		httpRequests := map[string]string{
			"assure_routes":             api.Routes,
			"assure_destinations":       api.Destinations,
			"assure_nodes":              api.Nodes,
			"assure_nodes_capabilities": api.NodesCapabilities,
			"assure_sms_routes":         api.SmsRoutes,
			"assure_sms_templates":      api.SmsTemplates,
		}
		keys := make([]string, 0)
		for key := range httpRequests {
			keys = append(keys, key)
		}
		for i := range keys {
			var err error
			res, err := api.newRequest("GET", api.QueryResults+httpRequests[keys[i]], nil)
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
		var route assureRoute // table's struct
		var routes routes      // JSON's struct
		if err := db.Delete(assureRoute{}).Error; err != nil {
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
		var destination assureDestination
		var destinations destinations
		if err := db.Delete(assureDestination{}).Error; err != nil {
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
		var nods nodes
		if err := db.Delete(assureNodes{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", node.TableName())
		if err := json.Unmarshal(body, &nods); err != nil {
			return err
		}
		for i := range nods.QueryResult1 {
			if err := mapstructure.Decode(nods.QueryResult1[i], &node); err != nil {
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
		var nodes nodeCapabilities
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
	case "assure_sms_routes":
		var route assureSmsRoute
		var routes smsRoutes
		if err := db.Delete(assureSmsRoute{}).Error; err != nil {
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
	case "assure_sms_templates":
		var template assureSmsTemplate
		var templates smsTemplates
		if err := db.Delete(assureSmsTemplate{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", template.TableName())
		if err := json.Unmarshal(body, &templates); err != nil {
			return err
		}
		for i := range templates.QueryResult1 {
			if err := mapstructure.Decode(templates.QueryResult1[i], &template); err != nil {
				return err
			}
			if dialectDB == "sqlite3" {
				if err := tx.Create(&template).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, template)
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

func (assureAPI) TableName() string {
	return schemaPG + "CallingSys_API_Assure"
}

type assureAPI struct {
	SystemName        string `gorm:"size:50"`
	SystemID          int    `gorm:"type:int"`
	URL               string `gorm:"size:100"`
	User              string `gorm:"size:100"`
	Pass              string `gorm:"size:100"`
	StatusTests       string `gorm:"size:100"`
	TestResults       string `gorm:"size:100"`
	QueryResults      string `gorm:"size:100"`
	Routes            string `gorm:"size:25"`
	Destinations      string `gorm:"size:25"`
	Nodes             string `gorm:"size:25"`
	NewTestGet        string `gorm:"size:25"`
	NodesCapabilities string `gorm:"size:25"`
	Version           string `gorm:"size:25"`
	SmsRoutes         string `gorm:"size:25"`
	SmsTemplates      string `gorm:"size:25"`
}

//-----------------------------------------------------------------------------
//*******************Block of Assure syncro structures*******************
//-----------------------------------------------------------------------------
type destinations struct {
	QueryResult1 []assureDestination
}

func (assureDestination) TableName() string {
	return schemaPG + "CallingSys_assure_destinations"
}

type assureDestination struct {
	CountryID                 int         `json:"CountryID" gorm:"type:int"`
	CountryName               string      `json:"CountryName" gorm:"type:varchar(100)"`
	Created                   string      `json:"Created" gorm:"type:varchar(100)"`
	CreatedBy                 int         `json:"CreatedBy" gorm:"type:int"`
	DestinationCategoryID     int         `json:"DestinationCategoryID" gorm:"type:int"`
	DestinationCategoryName   string      `json:"DestinationCategoryName" gorm:"type:varchar(100)"`
	DestinationExtID          interface{} `json:"DestinationExtID" gorm:"type:varchar(100)"`
	DestinationID             int         `json:"DestinationID" gorm:"type:int"`
	DestinationImportanceID   int         `json:"DestinationImportanceID"`
	DestinationImportanceName string      `json:"DestinationImportanceName" gorm:"type:varchar(100)"`
	Modified                  string      `json:"Modified" gorm:"type:varchar(100)"`
	ModifiedBy                int         `json:"ModifiedBy" gorm:"type:int"`
	Name                      string      `json:"Name" gorm:"type:varchar(100)"`
	PoPID                     int         `json:"PoPID" gorm:"type:int"`
	ShortName                 string      `json:"ShortName" gorm:"type:varchar(100)"`
}

type routes struct {
	QueryResult1 []assureRoute
}

func (assureRoute) TableName() string {
	return schemaPG + "CallingSys_assure_routes"
}

type assureRoute struct {
	Active                 bool        `json:"Active" gorm:"type:boolean"`
	CallParameter          interface{} `json:"CallParameter" gorm:"type:varchar(100)"`
	Carrier                string      `json:"Carrier" gorm:"type:varchar(100)"`
	CarrierID              interface{} `json:"CarrierID" gorm:"type:varchar(100)"`
	Channel                int         `json:"Channel" gorm:"type:int"`
	ChannelPoolID          int         `json:"ChannelPoolID" gorm:"type:int"`
	ChannelPoolName        string      `json:"ChannelPoolName" gorm:"type:varchar(100)"`
	Created                string      `json:"Created" gorm:"type:varchar(100)"`
	CreatedBy              int         `json:"CreatedBy" gorm:"type:int"`
	Description            interface{} `json:"Description" gorm:"type:varchar(100)"`
	DialerName             string      `json:"DialerName" gorm:"type:varchar(100)"`
	ExtRouteID             interface{} `json:"ExtRouteID" gorm:"type:varchar(100)"`
	IsMainProduct          interface{} `json:"IsMainProduct" gorm:"type:varchar(100)"`
	Modified               string      `json:"Modified" gorm:"type:varchar(100)"`
	ModifiedBy             int         `json:"ModifiedBy" gorm:"type:int"`
	Name                   string      `json:"Name" gorm:"type:varchar(100)"`
	NumberManipulationRule interface{} `json:"NumberManipulationRule" gorm:"type:varchar(100)"`
	PoPID                  int         `json:"PoPID" gorm:"type:int"`
	Prefix                 string      `json:"Prefix" gorm:"type:varchar(100)"`
	RouteClass             string      `json:"RouteClass" gorm:"type:varchar(100)"`
	RouteID                int         `json:"RouteID" gorm:"type:int"`
	RouteImportanceID      int         `json:"RouteImportanceID" gorm:"type:int"`
	RouteImportanceName    string      `json:"RouteImportanceName" gorm:"type:varchar(100)"`
	RouteTypeID            interface{} `json:"RouteTypeID" gorm:"type:varchar(100)"`
	RouteTypeName          interface{} `json:"RouteTypeName" gorm:"type:varchar(100)"`
	ShortName              string      `json:"ShortName" gorm:"type:varchar(100)"`
	SwitchID               int         `json:"SwitchID" gorm:"type:int"`
	SwitchName             string      `json:"SwitchName" gorm:"type:varchar(100)"`
}


type nodes struct {
	QueryResult1 []assureNodes
}

func (assureNodes) TableName() string {
	return schemaPG + "CallingSys_assure_nodes"
}

type assureNodes struct {
	APartyCallParameter     string `gorm:"size:150"`
	APartyCustomNumberMRule string `gorm:"size:100"`
	APartyNumberMRule       string `gorm:"size:150"`
	CallParameter           string `gorm:"size:100"`
	ChannelProperties       string `gorm:"size:100"`
	Created                 string `gorm:"size:100"`
	CreatedBy               int    `gorm:"type:int"`
	Description             string `gorm:"size:100"`
	DisplayName             string `gorm:"size:100"`
	EquipmentID             int    `gorm:"type:int"`
	EquipmentName           string `gorm:"size:100"`
	HomePLMN                string `gorm:"size:100"`
	IMEI                    string `gorm:"size:100"`
	InternalComment         string `gorm:"size:100"`
	IsSynchronized          bool   `gorm:"type:boolean"`
	ModeID                  int    `gorm:"type:int"`
	Modified                string `gorm:"size:100"`
	ModifiedBy              int    `gorm:"type:int"`
	ModifiedUTC             string `gorm:"size:100"`
	Name                    string `gorm:"size:100"`
	NetworkAndTestNodeName  string `gorm:"size:150"`
	NetworkName             string `gorm:"size:100"`
	NoNumberName            string `gorm:"size:100"`
	NoOfChannels            int    `gorm:"type:int"`
	PLMN                    string `gorm:"size:100"`
	PhoneNumber             string `gorm:"type:text"`
	PhoneNumberFirstPart    string `gorm:"size:100"`
	PhoneNumberIsEncrypted  bool   `gorm:"type:boolean"`
	ProbeContainerUID       string `gorm:"size:150"`
	RTDThreshold            string `gorm:"size:100"`
	TestNodeDirectoryID     int    `gorm:"type:int"`
	TestNodeDirectoryName   string `gorm:"size:100"`
	TestNodeID              int    `gorm:"type:int"`
	TestNodeUID             string `gorm:"size:150"`
}

type nodeCapabilities struct {
	QueryResult1 []assureNodesCapabilities
}

func (assureNodesCapabilities) TableName() string {
	return schemaPG + "CallingSys_assure_nodes_capabilities"
}

type assureNodesCapabilities struct {
	AAPoPID          int    `gorm:"type:int"`
	TestTypeName     string `gorm:"size:100"`
	DestinationIntID int    `gorm:"type:int"`
	DestinationID    string `gorm:"size:100"`
	DestinationName  string `gorm:"size:100"`
	TestNodeIntID    int    `gorm:"type:int"`
	TestNodeIntUID   string `gorm:"size:100"`
	TestNodeName     string `gorm:"size:100"`
	NetworkName      string `gorm:"size:100"`
	IsAParty         int    `gorm:"type:int"`
}

type smsRoutes struct {
	QueryResult1 []assureSmsRoute
}

func (assureSmsRoute) TableName() string {
	return schemaPG + "CallingSys_assure_sms_routes"
}

type assureSmsRoute struct {
	SMSRouteID             int         `json:"SMSRouteID" gorm:"type:int"`
	SMSRouteExtID          interface{} `json:"SMSRouteExtID" gorm:"type:varchar(100)"`
	PoPID                  int         `json:"PoPID" gorm:"type:int"`
	Name                   string      `json:"Name" gorm:"type:varchar(100)"`
	ShortName              string      `json:"ShortName" gorm:"type:varchar(25)"`
	Carrier                string      `json:"Carrier" gorm:"type:varchar(100)"`
	SMSAdapterInstanceID   int         `json:"SMSAdapterInstanceID" gorm:"type:int"`
	SMSAdapterInstanceName string      `json:"SMSAdapterInstanceName" gorm:"type:varchar(25)"`
	SMSRouteImportanceID   int         `json:"SMSRouteImportanceID" gorm:"type:int"`
	SMSRouteImportanceName string      `json:"SMSRouteImportanceName" gorm:"type:varchar(25)"`
	RouteClass             string      `json:"RouteClass" gorm:"type:varchar(25)"`
	Active                 bool        `json:"Active" gorm:"type:boolean"`
	SMSRouteProperties     string      `json:"SMSRouteProperties" gorm:"type:varchar(100)"`
	NumberManipulationRule interface{} `json:"NumberManipulationRule" gorm:"type:varchar(100)"`
	Description            interface{} `json:"Description" gorm:"type:varchar(100)"`
	RouteTypeID            interface{} `json:"RouteTypeID" gorm:"type:varchar(100)"`
	RouteTypeName          interface{} `json:"RouteTypeName" gorm:"type:varchar(100)"`
	IsMainProduct          interface{} `json:"IsMainProduct" gorm:"type:varchar(100)"`
	CreatedBy              int         `json:"CreatedBy" gorm:"type:int"`
	Created                string      `json:"Created" gorm:"type:varchar(100)"`
	ModifiedBy             int         `json:"ModifiedBy" gorm:"type:int"`
	Modified               string      `json:"Modified" gorm:"type:varchar(100)"`
}

type smsTemplates struct {
	QueryResult1 []assureSmsTemplate
}

func (assureSmsTemplate) TableName() string {
	return schemaPG + "CallingSys_assure_sms_templates"
}

type assureSmsTemplate struct {
	SMSTemplateID           int         `json:"SMSTemplateID" gorm:"type:int"`
	PoPID                   int         `json:"PoPID" gorm:"type:int"`
	Name                    string      `json:"Name" gorm:"type:varchar(100)"`
	SMSAdapterInstanceID    int         `json:"SMSAdapterInstanceID" gorm:"type:int"`
	SMSAdapaterInstanceName string      `json:"SMSAdapaterInstanceName" gorm:"type:varchar(25)"`
	TestTypeID              int         `json:"TestTypeID" gorm:"type:int"`
	TestTypeName            string      `json:"TestTypeName" gorm:"type:varchar(25)"`
	Description             interface{} `json:"Description" gorm:"type:varchar(100)"`
	CreatedBy               int         `json:"CreatedBy" gorm:"type:int"`
	Created                 string      `json:"Created" gorm:"type:varchar(100)"`
	ModifiedBy              int         `json:"ModifiedBy" gorm:"type:int"`
	Modified                string      `json:"Modified" gorm:"type:varchar(100)"`
}

//-----------------------------------------------------------------------------
// *******************Structs for initiated new tests*******************
//-----------------------------------------------------------------------------
type testSetDestination struct {
	NoOfExecutions int
	TestSetItems   []batchDestination
}

type batchDestination struct {
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

type testSetSMS struct {
	NoOfExecutions int
	TestSetItems   []batchSMS
}

type batchSMS struct {
	SMSRouteID      int
	DestinationID   int
	SMSTemplateName string
}

//`{"NoOfExecutions":%d,"TestSetItems": [{"SMSRouteID":%d,"DestinationID":%d,"SMSTemplateName":"%s"}]}

//-----------------------------------------------------------------------------
// ****************Structs for obtain tests status and results****************
//-----------------------------------------------------------------------------
type testStatusAssure struct {
	TestBatchID    int    `json:"TestBatchID"`
	StatusID       int    `json:"StatusID"`
	Status         string `json:"Status"`
	IsDone         bool   `json:"IsDone"`
	Created        string `json:"Created"` //time.Time
	TestBatchItems []struct {
		IsDone   bool   `json:"IsDone"`
		Status   string `json:"Status"`
		StatusID int    `json:"StatusID"`
	} `json:"TestBatchItems"`
}

func (tsa testStatusAssure) String() string {
	var status string
	switch tsa.StatusID {
	case 0:
		status = "Unknown"
	case 1:
		status = "Created"
	case 2:
		status = "Waiting"
	case 3:
		status = "Running"
	case 4:
		status = "Finishing"
	case 5:
		status = "Cancelling"
	case 6:
		status = "Cancelled"
	case 7:
		status = "Exception"
	}
	return status
}

type testResultAssure struct {
	QueryResult1 []batchResult
}

type batchResult struct {
	CallResultID            int
	ANumber                 string `json:"A number"`
	ANumberIsEncrypted      bool   `json:"A_NumberIsEncrypted"`
	BNumberIsEncrypted      bool   `json:"B_NumberIsEncrypted"`
	Result                  string
	TestType                string `json:"Test Type"`
	BNetwork                string `json:"B Network"`
	BTestNode               string `json:"B Test Node"`
	Bnumber                 string `json:"B number (first 6 digits)"`
	Route                   string
	TestStartTime           string `json:"Test Start Time"` //time.Time
	Alert                   interface{}
	Connect                 interface{}
	CLI                     interface{} //int
	VQA                     interface{} `json:"VQ A"`
	VQB                     interface{} `json:"VQ B"`
	FAS                     interface{} //int
	CLIDelivered            string      `json:"CLI Delivered"`
	MOSA                    float64     `json:"MOS A"`
	MOSB                    float64     `json:"MOS B"`
	PGRD                    float64
	PGAD                    float64
	CallDuration            float64     `json:"Call Duration"`
	DisconnectTime          float64     `json:"Disconnect Time"`
	BAlertTime              float64     `json:"B Alert Time"`
	BCallDuration           float64     `json:"B Call Duration"`
	BConnectTime            float64     `json:"B Connect Time"`
	BDisconnectTime         float64     `json:"B Disconnect Time"`
	ReleaseCause            string      `json:"Release Cause"`
	ReleaseLocation         string      `json:"Release Location"`
	SpeechFirstDetectedTime interface{} `json:"Speech First Detected Time"` //time.Time
	SpeechLastDetectedTime  interface{} `json:"Speech Last Detected Time"`  //time.Time
	APartyAudio             string      `json:"A Party Audio"`
}

type testResultAssureSMS struct {
	TestBatchResult []batchResultSMS `json:"TestBatchResult1"`
}
type batchResultSMS struct {
	CallBatchItemID           int         `json:"CallBatchItemID"`
	CallBatchStatusTypeID     interface{} `json:"CallBatchStatusTypeID"`
	CallBatchExceptionMessage interface{} `json:"CallBatchExceptionMessage"`
	UITestStatusID            int         `json:"UITestStatusID"`
	UITestStatusDisplay       string      `json:"UITestStatusDisplay"`
	Modified                  string      `json:"Modified"`
	HasAudio                  interface{} `json:"HasAudio"`
	HasSecondAudio            interface{} `json:"HasSecondAudio"`
	TestScenarioRuntimeUID    interface{} `json:"TestScenarioRuntimeUID"`
	SMSResultID               int         `json:"SMSResultID"`
	Result                    string      `json:"Result"`
	Network                   string      `json:"Network"`
	TestNode                  string      `json:"Test Node"`
	Supplier                  string      `json:"Supplier"`
	Route                     string      `json:"Route"`
	SentTime                  string      `json:"Sent Time"`
	DelTimeOK                 bool        `json:"Del. Time OK"`
	ContentOK                 bool        `json:"Content OK"`
	OAOK                      bool        `json:"OA OK"`
	STONOK                    bool        `json:"S. TON OK"`
	AlphaOK                   bool        `json:"Alpha OK"`
	DelRepOK                  bool        `json:"Del. Rep. OK"`
	DelTimeSec                int         `json:"Del. Time (sec)"`
	SMSC                      string      `json:"SMSC"`
	SMSCOwner                 string      `json:"SMSC Owner"`
	ErrorMsg                  string      `json:"Error Msg."`
	OAS                       string      `json:"OA (S)"`
	OAR                       string      `json:"OA (R)"`
	STONS                     string      `json:"S. TON (S)"`
	STONR                     string      `json:"S. TON (R)"`
	AlphaS                    string      `json:"Alpha (S)"`
	AlphaR                    string      `json:"Alpha (R)"`
	SNumPlanS                 interface{} `json:"S. Num. Plan (S)"`
	SNumPlanR                 string      `json:"S. Num. Plan (R)"`
	TextS                     string      `json:"Text (S)"`
	TextR                     string      `json:"Text (R)"`
	NoOfSegS                  int         `json:"No of Seg (S)"`
	NoOfSegR                  int         `json:"No of Seg (R)"`
	RecTime                   string      `json:"Rec. Time"`
	SMSCRecTime               string      `json:"SMSC Rec. Time"`
	DelRepTime                string      `json:"Del. Rep. Time"`
	MessageID                 int         `json:"Message Id"`
	Template                  string      `json:"Template"`
	SMS                       string      `json:"SMS"`
	DeliveryDetails           string      `json:"Delivery Details"`
	DelTimeLimit              string      `json:"Del. Time Limit"`
	Duplicates                int         `json:"Duplicates"`
	APIResponse               string      `json:"API Response"`
	ResultRecTime             string      `json:"Result Rec. Time"`
	UDHS                      interface{} `json:"UDH (S)"`
	UDHR                      interface{} `json:"UDH (R)"`
	PDUR                      string      `json:"PDU (R)"`
	ResultTrace               string      `json:"Result Trace"`
	ExceptionMsg              interface{} `json:"Exception Msg."`
	CustomerID                string      `json:"CustomerID"`
	SMSID                     string      `json:"SMSID"`
	SMSIDExpireTime           string      `json:"SMSIDExpireTime"`
	SMSTemplateID             int         `json:"SMSTemplateID"`
	SMSTemplateModified       interface{} `json:"SMSTemplateModified"`
	SubmitSMSResponseDelay    float64     `json:"SubmitSMSResponseDelay"`
	DeliveryUpdateDelay       float64     `json:"DeliveryUpdateDelay"`
	SMSReceiveDelay           float64     `json:"SMSReceiveDelay"`
	SMSResultReceiveID        int         `json:"SMSResultReceiveID"`
	SMSResultIDR              int         `json:"SMSResultIDR"`
	PoPIDR                    int         `json:"PoPIDR"`
	AdapterInstanceIDR        int         `json:"AdapterInstanceIDR"`
	AdapterInstanceNameR      string      `json:"AdapterInstanceNameR"`
	DCSCharacterSetR          string      `json:"DCSCharacterSetR"`
	SMR                       string      `json:"SMR"`
	SegmentDuplicate          int         `json:"SegmentDuplicate"`
	SMSResultSendID           int         `json:"SMSResultSendID"`
	SMSResultIDS              int         `json:"SMSResultIDS"`
	PoPIDS                    int         `json:"PoPIDS"`
	AdapterInstanceIDS        int         `json:"AdapterInstanceIDS"`
	AdapterInstanceNameS      string      `json:"AdapterInstanceNameS"`
	SubmitSMSResponseTime     string      `json:"SubmitSMSResponseTime"`
	StatusUpdateCode          string      `json:"StatusUpdateCode"`
	StatusUpdateTime          string      `json:"StatusUpdateTime"`
}

func (assureSMSResult) TableName() string {
	return schemaPG + "CallingSys_TestResultsAssureSMS"
}

type assureSMSResult struct {
	TestBatchID               string
	CallBatchItemID           int
	CallBatchStatusTypeID     interface{}
	CallBatchExceptionMessage interface{}
	UITestStatusID            int
	UITestStatusDisplay       string
	Modified                  time.Time
	HasAudio                  interface{}
	HasSecondAudio            interface{}
	TestScenarioRuntimeUID    interface{}
	SMSResultID               int
	Result                    string
	Network                   string
	TestNode                  string
	Supplier                  string
	Route                     string
	SentTime                  time.Time
	DelTimeOK                 bool
	ContentOK                 bool
	OAOK                      bool
	STONOK                    bool
	AlphaOK                   bool
	DelRepOK                  bool
	DelTimeSec                int
	SMSC                      string
	SMSCOwner                 string
	ErrorMsg                  string
	OAS                       string
	OAR                       string
	STONS                     string
	STONR                     string
	AlphaS                    string
	AlphaR                    string
	SNumPlanS                 interface{}
	SNumPlanR                 string
	TextS                     string
	TextR                     string
	NoOfSegS                  int
	NoOfSegR                  int
	RecTime                   time.Time
	SMSCRecTime               time.Time
	DelRepTime                time.Time
	MessageID                 int
	Template                  string
	SMS                       string
	DeliveryDetails           string
	DelTimeLimit              time.Time
	Duplicates                int
	APIResponse               string
	ResultRecTime             time.Time
	UDHS                      interface{}
	UDHR                      interface{}
	PDUR                      string
	ResultTrace               string
	ExceptionMsg              interface{}
	CustomerID                string
	SMSID                     string
	SMSIDExpireTime           time.Time
	SMSTemplateID             int
	SMSTemplateModified       interface{}
	SubmitSMSResponseDelay    float64
	DeliveryUpdateDelay       float64
	SMSReceiveDelay           float64
	SMSResultReceiveID        int
	SMSResultIDR              int
	PoPIDR                    int
	AdapterInstanceIDR        int
	AdapterInstanceNameR      string
	DCSCharacterSetR          string
	SMR                       string
	SegmentDuplicate          int
	SMSResultSendID           int
	SMSResultIDS              int
	PoPIDS                    int
	AdapterInstanceIDS        int
	SubmitSMSResponseTime     time.Time
	StatusUpdateCode          string
	StatusUpdateTime          time.Time
}
