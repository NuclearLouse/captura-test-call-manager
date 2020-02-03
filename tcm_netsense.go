// tcm_netsense.go
//
// The file contains the basic functions necessary for the operation of the "Netsense" system.
//
// Must be present functions that satisfy the tester interface
//
package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "redits.oculeus.com/asorokin/my_packages/logging"
)

func (api *netSenseAPI) sysName(db *gorm.DB) string {
	db.Take(api)
	return api.SystemName
}

func (api netSenseAPI) checkAuth(db *gorm.DB) bool {
	return true
}

func (api netSenseAPI) requestGET(r string) (*http.Response, error) {
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

func (api netSenseAPI) requestPOST(r string, xmlStr []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", api.URL+r, bytes.NewBuffer(xmlStr))
	req.Header.Set("Content-Type", "application/xml")
	if err != nil {
		return nil, err
	}
	res, err := api.httpRequest(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api netSenseAPI) httpRequest(req *http.Request) (*http.Response, error) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (netSenseAPI) parseBNumbers(customBNumbers string) (nums []string) {
	// Important:
	// Also	a phone	number can be used instead of a destination	code.
	// Plus(+) sign must be included.
	// Example:	+357123123123
	bnums := strings.Split(customBNumbers, "\n")
	for _, n := range bnums {
		if !strings.HasPrefix(n, "+") {
			n = "+" + n
		}
		nums = append(nums, n)
	}
	return
}

func netsenseRout(routID int) string {
	// обращение к базе по routID и возврат его имени
	return "route name"
}

func netsenseDestination(destID int) string {
	// обращение к базе по destID и возврат его имени
	return "+" + strconv.Itoa(destID)
}

func (api netSenseAPI) buildNewTest(nit foundTest) testInitNetsense {
	var (
		bnums []string
		dests []listDestination
	)
	if nit.BNumber != "" {
		bnums = api.parseBNumbers(nit.BNumber)
	}
	for i := 0; i < nit.TestCalls; i++ {
		dest := netsenseDestination(nit.DestinationID)
		if nit.BNumber != "" && bnums[i] != "" {
			dest = bnums[i]
		}
		destination := listDestination{Destination: dest}
		dests = append(dests, destination)
	}
	return testInitNetsense{
		Authentication: authentication{Key: api.AuthKey},
		ParametersList: parametersList{
			CallTypeList: callTypeList{
				List: listCallType{CallType: nit.TestType},
			},
			RouteList: routeList{
				ListRoute: listRoute{Route: netsenseRout(nit.TestSysRouteID)},
			},
			DestinationList: destinationList{List: dests},
		},
	}
}

func (api netSenseAPI) runNewTest(db *gorm.DB, nit foundTest) error {
	if nit.TestCalls == 0 {
		return errors.New("zero calls initialized")
	}

	newTest := api.buildNewTest(nit)

	xmlBody, err := xml.Marshal(newTest)
	if err != nil {
		return err

	}
	log.Debug("Build request body: ", string(xmlBody))
	response, err := api.requestPOST(api.TestInitList, xmlBody)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var newTests testInitResponse
	if err := xml.Unmarshal(body, &newTests); err != nil {
		return err
	}
	response.Body.Close()

	switch newTests.CallListResponseArray.Status.Code {
	case "200":
		testinfo := PurchOppt{
			TestingSystemRequestID: newTests.CallListResponseArray.ResponseID,
			RequestState:           2}
		// if err := db.Model(&newTestInfo).Where(`"RequestID"=?`, nit.RequestID).Update(newTestInfo).Error; err != nil {
		if err := testinfo.updateTestInfo(db, nit.RequestID); err != nil {
			return err
		}
	case "400":
		message := fmt.Sprintf("Bad Request. Error code:%s. %s\n", newTests.CallListResponseArray.SubStatus.List.Status.Code, newTests.CallListResponseArray.SubStatus.List.Status.Message)
		err := errors.New(message)
		testinfo := PurchOppt{
			TestingSystemRequestID: "0",
			TestedUntil:            time.Now(),
			TestComment:            message}
		// testinfo.failedTest(db, nit.RequestID, message)
		testinfo.updateTestInfo(db, nit.RequestID)
		return err
	}

	log.Info("Successful run test. TestID:", newTests.CallListResponseArray.ResponseID)
	return nil
}

func (api netSenseAPI) checkTestComplete(db *gorm.DB, lt foundTest) error {
	testid := lt.TestingSystemRequestID
	s := getStatus{
		CallList: testid,
	}
	xmlBody, err := xml.Marshal(s)
	if err != nil {
		return err
	}
	// fmt.Println(string(xmlBody))

	res, err := api.requestPOST(api.TestInitList, xmlBody)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// fmt.Println(string(body))
	var ts testStatusNetsense
	if err := xml.Unmarshal(body, &ts); err != nil {
		return err
	}
	var statistic PurchOppt
	switch ts.CallListResponseArray.CallLogResponses.Status {
	case "RUNNING":
		//TODO: Может нужен какой-то таймаут для запущенных тестов, некоторые могут висеть бесконечно.
		log.Debug("Wait. The test status is RUNNING for test_ID:", testid)
		return nil
	case "END":
		// начинаю забор результатов для ts.CallListResponseArray.CallLogResponses.CallListLogID
		log.Debug("The end test for test_ID", testid)
		request := fmt.Sprintf("%s/%s/1/%s", api.TestInit, api.AuthKey, ts.CallListResponseArray.CallLogResponses.CallListLogID)
		res, err := api.requestGET(request)
		log.Debugf("Sending request TestResults for system Assure test_ID %s", testid)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		log.Infof("Successful response to the request TestResults for system Assure test_ID %s", testid)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		//! If need, save response xml file for debug
		// if err := ioutil.WriteFile("test_"+testid+".xml", body, 0666); err != nil {
		// 	return err
		// }

		var callsinfo testResultNetsense
		if err := xml.Unmarshal(body, &callsinfo); err != nil {
			return err
		}
		start := time.Now()
		log.Debugf("Start transaction insert into the table TestResults for system Netsense test_id %s", testid)
		testedFrom, err := api.insertCallsInfo(db, callsinfo, lt)
		if err != nil {
			return err
		}
		log.Infof("Successfully insert data from table TestResults for system Netsense test_ID %s", testid)
		log.Debug("Elapsed time insert transaction", time.Since(start))

		statistic = callsStatistic(db, testid)
		statistic.TestedFrom = netsenseParseTime(testedFrom)
		statistic.TestedByUser = lt.RequestByUser
		statistic.TestResult = "OK"
		if err := statistic.updateStatistic(db, testid); err != nil {
			return err
		}
		log.Info("Successfully update data to the table Purch_Oppt from test_ID", testid)
		go api.uploadAudioFiles(db, callsinfo)
		return nil
	default:
		// ?Most likely this situation will never arise
		log.Info("Failed test for test_ID", testid)
		statistic.RequestState = 2
		statistic.TestedUntil = time.Now()
		statistic.TestComment = ts.CallListResponseArray.CallLogResponses.Status
		if err := statistic.updateStatistic(db, testid); err != nil {
			return err
		}
		log.Debug("Successfully update data to the table Purch_Oppt from test_ID", testid)
	}
	return nil
}

func (api netSenseAPI) insertCallsInfo(db *gorm.DB, tr testResultNetsense, lt foundTest) (string, error) {
	var testedFrom string
	for i := range tr.Result.List {
		if i == 0 {
			testedFrom = tr.Result.List[i].CallResult.CallStart.Value
		}
		res := tr.Result.List[i].CallResult
		resCLI := tr.Result.List[i].CallResultCLI
		resFAS := tr.Result.List[i].CallResultFAS
		callinfo := CallingSysTestResults{
			AudioURL:                 res.AudioURL.Value,
			CallID:                   res.CallResultID.Value,
			CallListID:               lt.TestingSystemRequestID,
			TestSystem:               lt.SystemID,
			CallType:                 res.CallType.Value,
			ConnectTime:              res.ConnectTime.Value,
			Destination:              res.Destination.Value,
			CallStart:                netsenseParseTime(res.CallStart.Value),
			CallComplete:             netsenseParseTime(res.CallComplete.Value),
			CallDuration:             res.CallDuration.Value,
			RingDuration:             res.RingingDuration.Value,
			PDD:                      res.Pdd.Value,
			BNumber:                  res.PhoneNumber.Value,
			BNumberDialed:            res.DialedPhoneNumber.Value,
			CallingNumber:            res.CallingNumber.Value,
			Route:                    res.Route.Value,
			CauseCodeID:              res.CauseCodeID.Value,
			CliDetectedCallingNumber: resCLI.CLIDetectedCallingNumber.Value,
			CliResult:                resCLI.CLIStatus.Value, //or resCLI.CLIResult.Value
			FasResult:                resFAS.FasResult.Value,
			Status:                   res.Status.Value,
		}
		if err := db.Create(&callinfo).Error; err != nil {
			return testedFrom, err
		}
	}
	return testedFrom, nil
}

func (api netSenseAPI) uploadAudioFiles(db *gorm.DB, tr testResultNetsense) {
	for _, l := range tr.Result.List {
		audioURL := l.CallResult.AudioURL.Value
		callID := l.CallResult.CallResultID.Value
		if audioURL == "" || strings.Split(audioURL, "/")[1] == "" {
			log.Info("Not present audio file for call_id", callID)
			if err := insertEmptyFiles(db, callID); err != nil {
				log.Errorf(401, "Cann't update data row about empty request for call_id %s|%v", callID, err)
			}
			continue
		}
		log.Infof("Download AudioURL:%s for call_id:%s", audioURL, callID)
		request := fmt.Sprintf("%s/%s/%s", api.AudioFile, api.AuthKey, audioURL)
		cWav, err := api.saveWavFile(request, callID)
		if err != nil {
			log.Errorf(402, "Cann't download the audio file for call_id:%s|%v", callID, err)
			continue
		}
		log.Info("Succeseffuly download audio file for call_id:", callID)

		x, _ := strconv.Atoi(fmt.Sprintf("%.f", 500*l.CallResult.ConnectTime.Value/(l.CallResult.ConnectTime.Value+l.CallResult.CallDuration.Value)))
		cImg, err := waveFormImage(callID, x)
		if err != nil || len(cImg) == 0 {
			log.Errorf(403, "Cann't create waveform image file for call_id %s|%v", callID, err)
			cImg = labelEmptyBMP("C&V:Cann't create waveform image file")
			// тут нужна проверка на очистку временной папки и вставка этой записи в таблицу
			continue
		}
		log.Info("Created image PNG file for call_id", callID)
		listDeleteFiles := []string{
			srvTmpFolder + callID + ".wav",
			srvTmpFolder + callID + ".png",
			srvTmpFolder + callID + ".bmp",
		}

		callsinfo := CallingSysTestResults{
			DataLoaded: true,
			AudioFile:  cWav,
			AudioGraph: cImg,
		}
		if err = callsinfo.updateCallsInfo(db, callID); err != nil {
			log.Errorf(404, "Cann't insert WAV file into table for system Assure call_id %s|%v", callID, err)
			continue
		}
		log.Info("Insert WAV and IMG file for callid", callID)

		if err = deleteFiles(listDeleteFiles); err != nil {
			log.Errorf(405, "Cann't delete some files for call_id %s|%v", callID, err)
		}
	}
}

func (api netSenseAPI) saveWavFile(request, nameFile string) ([]byte, error) {
	res, err := api.requestGET(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if strings.Contains(res.Header[`Content-Type`][0], "text/html") {
		return nil, errors.New("not present audio file, Content-Type:text/html")
	}
	if len(body) == 0 {
		return nil, errors.New("empty audio file")
	}
	if err := ioutil.WriteFile(srvTmpFolder+nameFile+".wav", body, 0666); err != nil {
		return nil, err
	}
	return body, nil
}
