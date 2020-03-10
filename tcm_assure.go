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

	log "./logger"

	"github.com/jinzhu/gorm"
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

func (assureAPI) newTestSMS(nit foundTest) testSetSMS {
	return testSetSMS{
		NoOfExecutions: nit.TestCalls,
		TestSetItems: []batchSMS{batchSMS{
			SMSRouteID:      nit.TestSysRouteID, // нужен route_id для SMS
			DestinationID:   nit.DestinationID,  // нужен destination_id для SMS
			SMSTemplateName: nit.TestComment},   // нужен template_name
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
		TestResult:             "Initiated",
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
	case 1, 2, 3:
		log.Trace("Wait. The test is not over yet for test_ID", testid)
	case 4:
		log.Info("The end test for test_ID", testid)

		var req string
		if strings.Contains(strings.ToLower(lt.TestType), "sms") == true {
			// Возможны два варианта запросов для получения результатов SMS тестов:
			//**************** 1 ********************
			// По ID теста
			// https://h-54-246-182-248.csg-assure.com/api/TestBatchResults/61825
			// req = api.TestResults + result.TestBatchID

			//**************** 2 ********************
			// По дате
			// https://h-54-246-182-248.csg-assure.com/api/QueryResults2?code=Details+:+SMS+MT&From=2020-02-24&To=2020-02-24
			req = fmt.Sprintf("%sDetails+:+SMS+MT&From=%s&To=%[2]s", api.QueryResults, time.Now().Format("2006-01-02"))
		} else {
			req = fmt.Sprintf("%sTest+Details+:+CLI+-+FAS+-+VQ+-+with+audio&Par1=%s", api.QueryResults, testid)
		}

		res, err := api.newRequest("GET", req, nil)
		log.Debugf("Sending request TestResults for system %s test_ID %s", api.SystemName, testid)
		if err != nil {
			return err
		}
		log.Infof("Successful response to the request TestResults for system %s test_ID %s", api.SystemName, testid)

		//! в структурe результатов теста надо добавить поля SMS
		var callsinfo testResultAssure
		if err := json.NewDecoder(res.Body).Decode(&callsinfo); err != nil {
			return err
		}
		res.Body.Close()

		start := time.Now()
		log.Debugf("Start transaction insert into the table TestResults for system Assure test_id %s", testid)
		//! При SMS тесте возможно надо вносить другие данные
		if err := api.insertCallsInfo(db, callsinfo, lt); err != nil {
			return err
		}
		log.Infof("Successfully insert data from table TestResults for system Assure test_ID %s", testid)
		log.Debug("Elapsed time insert transaction", time.Since(start))

		//! При SMS тесте статистику надо считать по другому, если вообще надо.
		ti.callsStatistic(db, testid)
		ti.TestedFrom = assureParseTime(result.Created)
		ti.TestedByUser = lt.RequestByUser
		// statistic.TestComment = "Bla bla bla test by Assure for test_ID:" + testid
		//! При SMS тесте не надо загружать аудио файлы
		go api.downloadAudioFiles(db, callsinfo)

	case 0, 5, 6, 7:
		log.Info("Cancelled test for test_ID", testid)
		ti.RequestState = 2
		ti.TestedUntil = time.Now()
		// statistic.TestComment = "Bla bla bla test by Assure for test_ID:" + testid
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
		callstart := assureParseTime(res.TestStartTime)
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
			Route:                    res.Route, // or lt.RouteCarrier
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
