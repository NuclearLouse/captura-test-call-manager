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
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "captura_tcm/logger"

	"github.com/mitchellh/mapstructure"

	"github.com/jinzhu/gorm"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
)

func (api *itestAPI) sysName(db *gorm.DB) string {
	db.Take(api)
	return api.SystemName
}

func (api *itestAPI) sysID(db *gorm.DB) int {
	db.Take(api)
	return api.SystemID
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

func (api itestAPI) cancelTest(db *gorm.DB, testid string) error {
	log.Debugf("Sending a request Cancel Test for system %s and test_id %s", api.SystemName, testid)

	return nil
}

func (api itestAPI) prepareRequests(db *gorm.DB, interval int64) {
	for {
		log.Info("Send preparatory requests for", api.SystemName)
		log.Debug("API Settings", api)
		httpRequests := map[string]int{
			"itest_profiles":      api.Profiles,
			"itest_suppliers":     api.Suppliers,
			"itest_breakouts_std": api.NdbStd,
			"itest_breakouts_cli": api.NdbCli,
		}
		keys := make([]string, 0)
		for key := range httpRequests {
			keys = append(keys, key)
		}
		for i := range keys {
			var err error
			req := fmt.Sprintf("%s?t=%d", api.URL, httpRequests[keys[i]])
			res, err := api.requestPOST(req)
			// log.Debug("Prepare response", response)
			if err != nil {
				log.Errorf(500, "Failed to get a response to the request %s|%v", keys[i], err)
				continue
			}
			log.Info("Successful response to the request", keys[i])
			start := time.Now()
			log.Debug("Start transaction insert into the table", keys[i])
			if err := insertsPrepareXML(db, keys[i], res); err != nil {
				log.Errorf(501, "Could not insert data from response %s|%v", keys[i], err)
				continue
			}
			log.Info("Successfully insert data from response", keys[i])
			log.Debugf("Elapsed time transaction insert %s %v", keys[i], time.Since(start))
		}
		log.Infof("The next data update to prepare %s after %d hours", api.SystemName, interval)
		time.Sleep(time.Duration(interval) * time.Hour)
	}

}

func insertsPrepareXML(db *gorm.DB, req string, res *http.Response) error {
	decoder := xmlNewDecoder(res)
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
	case "itest_profiles":
		var profile itestProfiles
		var profiles ProfilesList
		if err := db.Delete(itestProfiles{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", profile.TableName())
		if err := decoder.Decode(&profiles); err != nil {
			return err
		}
		for i := range profiles.Profiles {
			if err := mapstructure.Decode(profiles.Profiles[i], &profile); err != nil {
				return err
			}
			if dialectDB == "sqlite3" {
				if err := tx.Create(&profile).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, profile)
		}
	case "itest_suppliers":
		var supplier itestSuppliers
		var suppliers SuppliersList
		if err := db.Delete(itestSuppliers{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", supplier.TableName())
		if err := decoder.Decode(&suppliers); err != nil {
			return err
		}
		for i := range suppliers.Suppliers {
			if err := mapstructure.Decode(suppliers.Suppliers[i], &supplier); err != nil {
				return err
			}
			s := &supplier
			pref := strings.Split(s.Prefix, "#")
			s.Prefix = pref[0]
			if dialectDB == "sqlite3" {
				if err := tx.Create(&supplier).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, supplier)
		}
	default:
		var breakout itestBreakouts
		var ndblist ListNDB
		switch req {
		case "itest_breakouts_std":
			breakout = itestBreakouts{BreakType: "std"}
		case "itest_breakouts_cli":
			breakout = itestBreakouts{BreakType: "cli"}
		}
		if err := db.Table(breakout.TableName()).Delete(itestBreakouts{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", breakout.TableName())
		if err := decoder.Decode(&ndblist); err != nil {
			return err
		}
		for i := range ndblist.Breakouts {
			if err := mapstructure.Decode(ndblist.Breakouts[i], &breakout); err != nil {
				return err
			}
			if dialectDB == "sqlite3" {
				if err := tx.Create(&breakout).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, breakout)
		}
	}
	switch dialectDB {
	case "sqlite3":
		err := tx.Commit().Error
		if err != nil {
			return err
		}
	default:
		if err := gormbulk.BulkInsert(db, bulkslice, 3000); err != nil {
			return err
		}
	}
	return nil
}

func (api itestAPI) runNewTest(db *gorm.DB, nit foundTest) error {
	//! Default 5 calls for iTest
	if nit.TestCalls == 0 {
		return errors.New("zero calls initialized")
	}
	var request string
	// !All numerical values ​​are given as an example.
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

	res, err := api.requestPOST(request)
	if err != nil {
		return err
	}

	var ti testInitItest
	if err := xmlNewDecoder(res).Decode(&ti); err != nil {
		return err
	}
	res.Body.Close()

	testinfo := purchOppt{
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
	testid := lt.TestingSystemRequestID
	log.Debugf("Sending a request Complete_Test system %s for test_id %s", lt.SystemName, testid)
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

	log.Debugf("Successful response to the request Complete_Test for system %s test_ID %s", lt.SystemName, testid)
	newResp, err := ignoreWrongNode(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	var ts testStatusItest
	if err := xml.NewDecoder(newResp).Decode(&ts); err != nil {
		return err
	}

	var ti purchOppt
	// ! Тут нет проверки отмененного или зависшего теста, а именно смены ti.RequestState = 2
	ti.TestResult = "Running"
	switch ts.CallsTotal == ts.CallsComplete {
	case false:
		log.Info("Wait. The test is not over yet for test_ID", testid)
	case true:
		log.Info("The end test for test_ID", testid)
		req := fmt.Sprintf("%s?t=%d&jid=%s", api.URL, api.TestStatusDetails, testid)
		res, err := api.requestPOST(req)
		log.Debugf("Sending request TestResults fot system %s test_ID %s", lt.SystemName, testid)
		if err != nil {
			return err
		}
		log.Infof("Successful response to the request TestResults for system %s test_ID %s", lt.SystemName, testid)

		var tr testResultItest
		if err := xmlNewDecoder(res).Decode(&tr); err != nil {
			return err
		}
		res.Body.Close()

		start := time.Now()
		log.Debugf("Start transaction insert into the table TestResults for system %s test_id %s", lt.SystemName, testid)
		if err := api.insertCallsInfo(db, tr, lt); err != nil {
			return err
		}
		log.Infof("Successfully insert data from table TestResults for system %s test_ID %s", lt.SystemName, testid)
		log.Debug("Elapsed time insert transaction", time.Since(start))
		ti.callsStatistic(db, testid)
		ti.TestedFrom = time.Unix(tr.TestOverview.Init, 0)
		ti.TestResult = "Finishing"

		go api.downloadAudioFiles(db, tr)
	}

	if err := ti.updateStatistic(db, testid); err != nil {
		return err
	}
	log.Debug("Successfully update data to the table Purch_Oppt from test_ID", testid)
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
		cMP3, err := contentMP3(nameFile)
		if err != nil || len(cMP3) == 0 {
			log.Errorf(504, "Cann't read MP3 file for call_id %s|%v", call.ID, err)
		}

		cWav, err := decodeAudio(nameFile, "mp3", "wav", "source")
		if err != nil || len(cWav) == 0 {
			log.Errorf(505, "Cann't decode MP3 file to WAV for call_id %s|%v", call.ID, err)
			continue
		}
		log.Info("Decod mp3 files to WAV for call_id", call.ID)

		pngImg, bmpImg, err := waveFormImage(nameFile, x)
		if err != nil || len(bmpImg) == 0 || len(pngImg) == 0 {
			log.Errorf(506, "Cann't create waveform image files for call_id %s|%v", call.ID, err)
			bmpImg = labelEmptyBMP("C&V:Cann't create waveform image file")
			continue
		}
		log.Info("Created image files for call_id", call.ID)

		callsinfo := callingSysTestResults{
			DataLoaded:  true,
			AudioFile:   cWav,
			AudioGraph:  bmpImg,
			ConnectTime: connectTime,
		}
		if err = callsinfo.updateCallsInfo(db, call.ID); err != nil {
			log.Errorf(507, "Cann't insert audio and image file into TestResults table for call_id %s|%v", call.ID, err)
			continue
		}

		webinfo := testFilesWEB{
			Callid:     call.ID,
			Testsystem: api.SystemID,
			Diagram:    pngImg,
			Audiofile:  cMP3}
		if err = webinfo.insertWebInfo(db); err != nil {
			log.Errorf(508, "Cann't insert audio and image file into testfiles_web table for call_id %s|%v", call.ID, err)
			continue
		}

		log.Info("Insert audio and image file for callid", call.ID)

		// if err := os.Remove(nameFile + ".mp3"); err != nil {
		// 	log.Error(4, "Cann't delete file", nameFile)
		// }
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
		callinfo := callingSysTestResults{
			AudioURL:                 ti.TestComment,
			CallID:                   call.ID,
			CallListID:               tr.TestOverview.TestID,
			TestSystem:               ti.SystemID,
			CallType:                 ti.TestType,
			Destination:              call.Destination,
			CallStart:                time.Unix(call.Start, 0),
			CallComplete:             time.Unix(int64(call.End), 0),
			CallDuration:             callDuration,
			RingDuration:             ringDuration,
			PDD:                      call.PDD,
			BNumber:                  prefixAndBnumber[1],
			BNumberDialed:            call.Destination,
			CallingNumber:            call.Source,
			Route:                    tr.TestOverview.Supplier,
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
	SystemName        string `gorm:"size:50"`
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

// ListNDB ...
type ListNDB struct {
	XMLName   xml.Name   `xml:"NDB_List"`
	Breakouts []breakout `xml:"Breakout"`
}

type breakout struct {
	XMLName      xml.Name `xml:"Breakout"`
	CountryName  string   `xml:"Country_Name"`
	CountryID    string   `xml:"Country_ID"`
	BreakoutName string   `xml:"Breakout_Name"`
	BreakoutID   string   `xml:"Breakout_ID"`
}

// ProfilesList ...
type ProfilesList struct {
	XMLName  xml.Name  `xml:"Profiles_List"`
	Profiles []profile `xml:"Profile"`
}

type profile struct {
	XMLName          xml.Name `xml:"Profile"`
	ProfileID        string   `xml:"Profile_ID"`
	ProfileName      string   `xml:"Profile_Name"`
	ProfileIP        string   `xml:"Profile_IP"`
	ProfilePort      string   `xml:"Profile_Port"`
	ProfileSrcNumber string   `xml:"Profile_Src_Number"`
}

// SuppliersList ....
type SuppliersList struct {
	XMLName   xml.Name   `xml:"Vendors_List"`
	Suppliers []supplier `xml:"Supplier"`
}

type supplier struct {
	XMLName      xml.Name `xml:"Supplier"`
	SupplierID   string   `xml:"Supplier_ID"`
	SupplierName string   `xml:"Supplier_Name"`
	Prefix       string   `xml:"Prefix"`
	Codec        string   `xml:"Codec"`
}

func (itestProfiles) TableName() string {
	return schemaPG + "CallingSys_iTest_profiles"
}

type itestProfiles struct {
	ProfileID        string `gorm:"column:profile_id;size:100"`
	ProfileName      string `gorm:"column:profile_name;size:100"`
	ProfileIP        string `gorm:"column:profile_ip;size:100"`
	ProfilePort      string `gorm:"column:profile_port;size:100"`
	ProfileSrcNumber string `gorm:"column:profile_src_number;size:100"`
}

func (itestSuppliers) TableName() string {
	return schemaPG + "CallingSys_iTest_suppliers"
}

type itestSuppliers struct {
	SupplierID   string `gorm:"column:supplier_id;size:100"`
	SupplierName string `gorm:"column:supplier_name;size:100"`
	Prefix       string `gorm:"column:prefix;size:100"`
	Codec        string `gorm:"column:codec;size:100"`
}

func (b itestBreakouts) TableName() string {
	var name string
	switch b.BreakType {
	case "cli":
		name = schemaPG + "CallingSys_iTest_breakouts_cli"
	case "std":
		name = schemaPG + "CallingSys_iTest_breakouts_std"
	}
	return name
}

type itestBreakouts struct {
	CountryName  string `gorm:"column:country_name;size:100"`
	CountryID    string `gorm:"column:country_id;size:100"`
	BreakoutName string `gorm:"column:breakout_name;size:100"`
	BreakoutID   string `gorm:"column:breakout_id;size:100"`
	BreakType    string `gorm:"-"`
}

