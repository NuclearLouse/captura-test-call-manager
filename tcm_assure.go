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
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "redits.oculeus.com/asorokin/CaptTestCallsSrvc/logger"
)

func (api *assureAPI) sysName(db *gorm.DB) string {
	db.Take(api)
	return api.SystemName
}

func (api assureAPI) checkAuth(db *gorm.DB) bool {
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
			PhoneNumber:  p}

		batches = append(batches, b)
	}
	return testSetBnumbers{
		TestSetItems: batches,
	}
}

func (assureAPI) newTestDestination(nit foundTest) testSetDestination {
	return testSetDestination{
		NoOfExecutions: nit.TestCalls,
		TestSetItems: []batchDestination{batchDestination{
			TestTypeName:  nit.TestType,
			RouteID:       nit.TestSysRouteID,
			DestinationID: nit.DestinationID},
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
	return api.newTestDestination(nit), nil
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

	newTest, err := api.buildNewTests(nit)
	if err != nil {
		return err
	}
	jsonBody, err := json.Marshal(newTest)
	if err != nil {
		return err
	}
	log.Debug("Build request body: ", string(jsonBody))
	res, err := api.requestPOST(api.StatusTests, jsonBody)
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

	testinfo := PurchOppt{
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
	res, err := api.requestGET(api.StatusTests + testid)
	if err != nil {
		return err
	}
	log.Debugf("Successful response to the request Complete_Test for system %s test_ID %s", api.SystemName, testid)

	var result testStatusAssure
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return err
	}
	res.Body.Close()

	var statistic PurchOppt
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
		//Test Details : CLI - FAS - VQ - with audio
		req := fmt.Sprintf("%sTest+Details+:+CLI+-+FAS+-+VQ+-+with+audio&Par1=%s", api.QueryResults, testid)
		res, err := api.requestGET(req)
		log.Debugf("Sending request TestResults for system %s test_ID %s", api.SystemName, testid)
		if err != nil {
			return err
		}
		log.Infof("Successful response to the request TestResults for system %s test_ID %s", api.SystemName, testid)

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
		statistic = callsStatistic(db, testid)
		statistic.TestedFrom = assureParseTime(result.Created)
		statistic.TestedByUser = lt.RequestByUser
		statistic.TestResult = "OK"
		if err := statistic.updateStatistic(db, testid); err != nil {
			return err
		}
		log.Info("Successfully update data to the table Purch_Oppt from test_ID", testid)
		go api.downloadAudioFiles(db, callsinfo)
		return nil
	case 5, 6, 7:
		// 5 - Cancelling
		// 6 - Cancelled
		// 7 - Exception
		log.Info("Cancelled test for test_ID", testid)
		statistic.RequestState = 2
		statistic.TestedUntil = time.Now()
		statistic.TestComment = "Cancelled test by Assure for test_ID:" + testid
		if err := statistic.updateStatistic(db, testid); err != nil {
			return err
		}
		log.Debug("Successfully update data to the table Purch_Oppt from test_ID", testid)
		return nil
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
		var cWav []byte
		cWav, err = decodeToWAV(callID, "amr")
		if err != nil || len(cWav) == 0 {
			log.Errorf(608, "Error decode to wav file for call_id %s|%v", callID, err)
			cWav = []byte("C&V:Cann't decode to wav file")
			continue
		}
		log.Info("Created WAV file for call_id", callID)

		x, _ := strconv.Atoi(fmt.Sprintf("%.f", 500*res.BConnectTime/res.BDisconnectTime))

		cImg, err := waveFormImage(callID, x)
		if err != nil || len(cImg) == 0 {
			log.Errorf(609, "Cann't create waveform image file for call_id %s|%v", callID, err)
			cImg = labelEmptyBMP("C&V:Cann't create waveform image file")
			continue
		}
		log.Info("Created image PNG file for call_id", callID)

		callsinfo := CallingSysTestResults{
			DataLoaded: true,
			AudioFile:  cWav,
			AudioGraph: cImg,
		}
		if err = callsinfo.updateCallsInfo(db, callID); err != nil {
			log.Errorf(610, "Cann't insert WAV file into table for system Assure call_id %s|%v", callID, err)
			continue
		}
		log.Info("Insert WAV and IMG file for callid", callID)
	}
}

func (assureAPI) insertCallsInfo(db *gorm.DB, tr testResultAssure, lt foundTest) error {
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
