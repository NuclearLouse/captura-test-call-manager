// tcm_itest.go
//
// The file contains the functions necessary for the operation of the "iTest" system.
//
// Must be present functions that satisfy the tester interface
//
package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "redits.oculeus.com/asorokin/my_packages/logging"
)

func (api *itestAPI) sysName(db *gorm.DB) string {
	db.Take(api)
	return api.SystemName
}

func (api itestAPI) checkAuth(db *gorm.DB) bool {
	return true
}

func (api itestAPI) requestPOST(req string) (*http.Response, error) {
	res, err := http.PostForm(req, url.Values{"email": {api.User}, "pass": {api.Pass}})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api itestAPI) parseBNumbers(customBNumbers string) (nums string) {
	bnums := strings.Split(customBNumbers, "\n")
	for _, n := range bnums {
		nums = nums + strings.TrimPrefix(n, "+") + "%%-"
	}
	return
}

func (api itestAPI) runNewTest(db *gorm.DB, nit foundTest) error {
	//! Default 5 calls for iTest
	if nit.TestCalls == 0 {
		return errors.New("zero calls initialized")
	}
	// request := api.buildNewTest(nit)
	var request string
	var profileID int = 8524 //will get from nit, this is profile_id from CallingSys_iTest_profiles
	num := nit.TestCalls
	if nit.BNumber != "" {
		prefix := 216215 //will get from nit, prefix from CallingSys_iTest_suppliers
		request = fmt.Sprintf("%s?t=%d&profid=%d&codec=g729&prefix=%d&numbers=%s",
			api.URL, api.TestInit, profileID, prefix, api.parseBNumbers(nit.BNumber))
	}
	countryID := 432   //from CallingSys_iTest_breakouts_cli or CallingSys_iTest_breakouts_std
	breakoutID := 2517 //from CallingSys_iTest_breakouts_cli or CallingSys_iTest_breakouts_std
	switch strings.ToLower(nit.TestType) {
	case "cli":
		vended := 567349 //nit.SupplierID after sinchronization, supplier_id from CallingSys_iTest_suppliers
		request = fmt.Sprintf("%s?t=%d&profid=%d&vended=%d&ndbccgid=%d&ndbcgid=%d&qty=%d",
			api.URL, api.TestInitCli, profileID, vended, countryID, breakoutID, num)
	case "voice":
		prefix := 216215 //will get from nit, prefix from CallingSys_iTest_suppliers
		request = fmt.Sprintf("%s?t=%d&profid=%d&prefix=%d&ndbccgid=%d&ndbcgid=%d&qty=%d",
			api.URL, api.TestInit, profileID, prefix, countryID, breakoutID, num)
	}

	response, err := api.requestPOST(request)
	if err != nil {
		return err
	}
	decoder := xmlDecoder(response)
	var ti testInitItest
	if err := decoder.Decode(&ti); err != nil {
		return err
	}
	testinfo := PurchOppt{
		TestingSystemRequestID: ti.Test.TestID,
		TestComment:            ti.Test.ShareURL,
		RequestState:           2}
	if err := testinfo.updateTestInfo(db, nit.RequestID); err != nil {
		return err
	}
	log.Infof("Successful run test. TestID:%s", ti.Test.TestID)
	return nil
}

func (api itestAPI) checkTestComplete(db *gorm.DB, lt foundTest) error {
	sysname := lt.SystemName //or api.sysName()
	testid := lt.TestingSystemRequestID
	log.Debugf("Sending a request Complete_Test system %s for test_id %s", sysname, testid)
	req := fmt.Sprintf("%s?t=%d&jid=%s", api.URL, api.TestStatus, testid)
	res, err := api.requestPOST(req)
	log.Debug("Complete Test response", res)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			log.Info("Wait. The test has not yet begun for test_ID", testid)
		}
	}()

	log.Debugf("Successful response to the request Complete_Test for system %s test_ID %s", sysname, testid)
	newResp, err := ignoreWrongNode(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()
	body, err := ioutil.ReadAll(newResp)
	if err != nil {
		return err
	}

	var s testStatusItest
	if err := xml.Unmarshal(body, &s); err != nil {
		return err
	}

	if s.CallsTotal == s.CallsComplete {
		req := fmt.Sprintf("%s?t=%d&jid=%s", api.URL, api.TestStatusDetails, testid)
		res, err := api.requestPOST(req)
		log.Debugf("Sending request TestResults fot system %s test_ID %s", sysname, testid)
		if err != nil {
			return err
		}
		log.Infof("Successful response to the request TestResults for system %s test_ID %s", sysname, testid)
		decoder := xmlDecoder(res)
		var tr testResultItest
		if err := decoder.Decode(&tr); err != nil {
			return err
		}
		start := time.Now()
		log.Debugf("Start transaction insert into the table TestResults for system %s test_id %s", sysname, testid)
		if err := api.insertCallsInfo(db, tr, lt); err != nil {
			return err
		}
		log.Infof("Successfully insert data from table TestResults for system %s test_ID %s", sysname, testid)
		log.Debug("Elapsed time insert transaction", time.Since(start))
		statistic := callsStatistic(db, testid)
		statistic.TestedFrom = time.Unix(tr.TestOverview.Init, 0)
		if err := statistic.updateStatistic(db, testid); err != nil {
			return err
		}
		log.Info("Successfully update data to the table Purch_Oppt from test_ID", testid)
		go api.downloadAudioFiles(db, tr)
		return nil
	}
	log.Info("Wait. The test is not over yet for test_ID", testid)
	return nil
}

func (api itestAPI) checkPresentAudioFile(req string) (bool, error) {
	res, err := api.requestPOST(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.Header[`Content-Type`][0] != "audio/mpeg" {
		return false, errors.New("not present audio file")
	}
	//"http://share.i-test.net/au/20200203/call-2020020315000005864-r.mp3"
	//"http://share.i-test.net/au/20200203/call-2020020315000005864.mp3"
	//call-20200203123456789-r.mp3 or call-20200203123456789.mp3
	nameFile := strings.Split(req, "/")[5]
	if err := createFile(res.Body, nameFile); err != nil {
		return false, err
	}
	return true, nil
}

func (api itestAPI) downloadAudioFiles(db *gorm.DB, tr testResultItest) {
	for _, call := range tr.Call {
		req := fmt.Sprintf("%s%s/call-%s-r.mp3", api.RepoURL, call.ID[:8], call.ID)
		fileRing, err := api.checkPresentAudioFile(req)
		if err != nil {
			switch err.Error() {
			case "not present audio file":
				log.Infof("Test %v 'ring' for callID:%s", err, call.ID)
			default:
				log.Errorf(502, "Error in 'ring'checkPresentAudioFile function for callID:%s|%v", call.ID, err)
			}
		}
		req = fmt.Sprintf("%s%s/call-%s.mp3", api.RepoURL, call.ID[:8], call.ID)
		fileAnsw, err := api.checkPresentAudioFile(req)
		if err != nil {
			switch err.Error() {
			case "not present audio file":
				log.Infof("Test %v 'answer' for callID:%s", err, call.ID)
			default:
				log.Errorf(502, "Error in 'answer'checkPresentAudioFile function for callID:%s|%v", call.ID, err)
			}
		}
		var nameFile string
		var cWav, cImg []byte
		var connectTime float64
		var x int
		switch {
		case fileRing && fileAnsw:
			connectTime, x, err = calcCoordinate(call.ID)
			if err != nil {
				log.Errorf(509, "Cann't get the coordinate of the beginning of the answer for call_ID %s|%v", call.ID, err)
			}
			if err := concatMP3files(call.ID); err != nil {
				log.Errorf(510, "Cann't concatenate ring and answer MP3 files for call_id %s|%v", call.ID, err)
				continue
			}
			log.Info("Concatenate ring and answer mp3 files for call_id", call.ID)

			nameFile = "out-" + call.ID
		case fileRing && !fileAnsw:
			nameFile = "call-" + call.ID + "-r"
		case !fileRing && fileAnsw:
			nameFile = "call-" + call.ID
		case !fileRing && !fileAnsw:
			if err = insertEmptyFiles(db, call.ID); err != nil {
				log.Errorf(503, "Cann't update data row about empty request for call_id %s|%v", call.ID, err)
			}
			continue
		}
		cWav, err = decodeToWAV(nameFile, "mp3")
		if err != nil {
			log.Errorf(505, "Cann't decode MP3 file to WAV for call_id %s|%v", call.ID, err)
			continue
		}
		log.Info("Decod mp3 files to WAV for call_id", call.ID)
		cImg, err = waveFormImage(nameFile, x)
		if err != nil || len(cImg) == 0 {
			log.Errorf(506, "Cann't create waveform PNG image file for call_id %s|%v", call.ID, err)
			cImg = labelEmptyBMP("C&V:Cann't create waveform image file")
			continue
		}
		log.Info("Created image PNG file for call_id", call.ID)

		callsinfo := CallingSysTestResults{
			DataLoaded:  true,
			AudioFile:   cWav,
			AudioGraph:  cImg,
			ConnectTime: connectTime,
		}
		if err = callsinfo.updateCallsInfo(db, call.ID); err != nil {
			log.Errorf(507, "Cann't insert WAV file into table for call_id %s|%v", call.ID, err)
			continue
		}
		log.Info("Insert WAV and IMG file for callid", call.ID)

		if err := os.Remove(nameFile); err != nil {
			log.Error(4, "Cann't delete file", nameFile)
		}
	}
	log.Infof("All present files download from %s server and upload into the table TestResults", api.SystemName)
}

func (itestAPI) insertCallsInfo(db *gorm.DB, tr testResultItest, ti foundTest) error {
	var err error
	var ringDuration, callDuration float64
	for _, call := range tr.Call {
		prefixAndBnumber := strings.Split(call.Destination, "#")
		switch call.Call {
		case "NA":
			callDuration = -1
		default:
			callDuration, err = strconv.ParseFloat(call.Call, 64)
			if err != nil {
				callDuration = -1
			}
		}
		switch call.Ring {
		case "NA":
			ringDuration = -1
		default:
			ringDuration, err = strconv.ParseFloat(call.Ring, 64)
			if err != nil {
				ringDuration = -1
			}
		}
		callinfo := CallingSysTestResults{
			AudioURL:                 ti.TestComment,
			CallID:                   call.ID,
			CallListID:               tr.TestOverview.TestID, //ti.TestingSystemRequestID,
			TestSystem:               ti.SystemID,
			CallType:                 ti.TestType, //tr.TestOverview.Type,
			Destination:              call.Destination,
			CallStart:                time.Unix(call.Start, 0),
			CallComplete:             time.Unix(int64(call.End), 0),
			CallDuration:             callDuration,
			RingDuration:             ringDuration,
			PDD:                      call.PDD,
			BNumber:                  prefixAndBnumber[1],
			BNumberDialed:            call.Destination,
			CallingNumber:            call.Source,
			Route:                    tr.TestOverview.Supplier, //ti.RouteCarrier,
			CauseCodeID:              call.ResultCode,
			Status:                   call.LastCode,
			CliDetectedCallingNumber: call.CLI,
			CliResult:                call.Result,
			FasResult:                call.FAS,
			VoiceQualityMos:          call.MOS,
		}

		if err = db.Create(&callinfo).Error; err != nil {
			return err
		}
	}
	return nil
}

func (itestAPI) TableName() string {
	return schemaPG + "CallingSys_API_iTest"
}

type itestAPI struct {
	SystemName        string `gorm:"size:50;foreignkey:CallingSys_Settings.SystemName"`
	URL               string `gorm:"size:100"`
	RepoURL           string `gorm:"size:100"`
	User              string `gorm:"size:100"`
	Pass              string `gorm:"size:100"`
	Profiles          int    `gorm:"type:int"`
	Suppliers         int    `gorm:"type:int"`
	NdbStd            int    `gorm:"type:int"`
	NdbCli            int    `gorm:"type:int"`
	TestInit          int    `gorm:"type:int"`
	TestInitCli       int    `gorm:"type:int"`
	TestStatus        int    `gorm:"type:int"`
	TestStatusDetails int    `gorm:"type:int"`
	SystemID          int    `gorm:"type:int"`
}

type testInitItest struct {
	XMLName xml.Name `xml:"Test_Initiation"`
	Test    struct {
		TestID   string `xml:"Test_ID"`
		ShareURL string `xml:"Share_URL"`
	} `xml:"Test"`
}

type testStatusItest struct {
	XMLName       xml.Name `xml:"Test_Status"`
	Name          string   `xml:"Name"`
	CallsTotal    int      `xml:"Calls_Total"`
	CallsComplete int      `xml:"Calls_Complete"`
	CallsSuccess  int      `xml:"Calls_Success"`
	CallsNoAnswer int      `xml:"Calls_No_Answer"`
	CallsFail     int      `xml:"Calls_Fail"`
	PDD           float64  `xml:"PDD"`
	ShareURL      string   `xml:"Share_URL"`
}

type testResultItest struct {
	XMLName      xml.Name `xml:"Test_Status"`
	TestOverview struct {
		Name     string `xml:"Name"`
		Supplier string `xml:"Supplier"`
		InitBy   string `xml:"InitBy"`
		Init     int64  `xml:"Init"`
		Type     string `xml:"Type"`
		TestID   string `xml:"Test_ID"`
	} `xml:"Test_Overview"`
	Call []struct {
		ID          string  `xml:"ID"`
		Source      string  `xml:"Source"`
		Destination string  `xml:"Destination"`
		Start       int64   `xml:"Start"`
		End         float64 `xml:"End"`
		PDD         float64 `xml:"PDD"`
		MOS         float64 `xml:"MOS"`
		Ring        string  `xml:"Ring"`
		Call        string  `xml:"Call"`
		LastCode    string  `xml:"Last_Code"`
		Result      string  `xml:"Result"`
		ResultCode  int     `xml:"Result_Code"`
		CLI         string  `xml:"CLI"`
		FAS         string  `xml:"FAS"`
		LDFAS       string  `xml:"LD_FAS"`
		DeadAir     string  `xml:"Dead_Air"`
		NoRBT       string  `xml:"No_RBT"`
		Viber       string  `xml:"Viber"`
		FDLR        string  `xml:"F-DLR"`
	} `xml:"Call"`
}
