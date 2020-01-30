// netsense.go
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

func (api netSenseAPI) prepareRequests(db *gorm.DB, interval int64) {
	log.Info("Send preparatory requests for", api.SystemName)
	log.Debug("API Settings", api)
	for {
		log.Infof("The next data update to prepare %s after %d hours", api.SystemName, interval)
		time.Sleep(time.Duration(interval) * time.Hour)
	}

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
	return "destination name"
}

func (api netSenseAPI) newTest(ttn string, nit foundTest) testInit {
	var (
		routes    []routeList
		destes    []destList
		calltypes []typeList
		bnums     []string
	)
	if nit.BNumber != "" {
		bnums = api.parseBNumbers(nit.BNumber)
	}

	for i := 0; i < nit.TestCalls; i++ {
		dest := netsenseDestination(nit.DestinationID)
		if nit.BNumber != "" && bnums[i] != "" {
			dest = bnums[i]
		}
		r := routeList{Route: netsenseRout(nit.TestSysRouteID)}
		d := destList{Destination: dest}
		t := typeList{CallType: ttn}
		routes = append(routes, r)
		destes = append(destes, d)
		calltypes = append(calltypes, t)
	}
	return testInit{
		Auth: auth{Key: api.AuthKey},
		Parameters: parameters{
			RoutesList: routesList{
				List: routes,
			},
			DestinationsList: destinationsList{
				List: destes,
			},
			CallTypeList: callTypeList{
				List: calltypes,
			},
		},
		// Default value
		Settings: settings{
			TimeZone:     nit.TimeZone,     //"Europe/Stockholm"
			WebServiceID: nit.WebServiceID, //1
		},
	}
}

func (api netSenseAPI) buildNewTests(ttn string, nit foundTest) (testInit, error) {
	if nit.TestCalls == 0 {
		return testInit{}, errors.New("zero calls initialized")
	}
	return api.newTest(ttn, nit), nil
}

func (api netSenseAPI) runNewTest(db *gorm.DB, nit foundTest) error {
	if nit.TestCalls == 0 {
		return errors.New("zero calls initialized")
	}

	var ttn string
	switch nit.TestType.name() {
	case "cli":
		ttn = "TrueCLI"
	case "voice":
		ttn = "Voice"
	}

	newTest := api.newTest(ttn, nit)

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

	var newTests responseTestInit
	if err := xml.Unmarshal(body, &newTests); err != nil {
		return err
	}
	response.Body.Close()

	log.Debug(string(body))

	// if newTests.TestBatchID == 0 {
	// 	err := errors.New("no return TestingSystemRequestID")
	// 	testinfo := PurchOppt{TestingSystemRequestID: "0"}
	// 	testinfo.failedTest(db, nit.RequestID, string(body))
	// return err
	// }

	// newTestInfo := PurchOppt{
	// 	TestingSystemRequestID: strconv.Itoa(newTests.TestBatchID),
	// 	RequestState:           2}
	// if err := db.Model(&newTestInfo).Where(`"RequestID"=?`, nit.RequestID).Update(newTestInfo).Error; err != nil {
	//return err
	// }
	// log.Infof("Successful run test. TestID:%d", newTests.TestBatchID)

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
	var ts testStatus
	if err := xml.Unmarshal(body, &ts); err != nil {
		return err
	}
	var statistics PurchOppt
	switch ts.CallListResponseArray.CallLogResponses.Status {
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
		log.Infof("Successful response to the request TestResults for system %s test_ID %s", sysname, testid)
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		var callsinfo testResult
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

		statistics = callsStatistics(db, testid)
		statistics.TestedFrom = netsenseParseTime(testedFrom)
		statistics.TestedByUser = lt.RequestByUser
		statistics.TestResult = "OK"
		if err = db.Model(&statistics).Where(`"TestingSystemRequestID"=?`, testid).Update(statistics).Error; err != nil {
			return err
		}
		log.Info("Successfully update data to the table Purch_Oppt from test_ID", testid)
		go api.checkPresentAudioFile(db, callsinfo)
		return nil
	case "RUNNING":
		log.Debug("Wait. The test is not over yet for test_ID", testid)
		return nil
	default:
		log.Info("Failed test for test_ID", testid)
		statistics.RequestState = 2
		statistics.TestedUntil = time.Now()
		statistics.TestComment = ts.CallListResponseArray.CallLogResponses.Status
		if err = db.Model(&statistics).Where(`"TestingSystemRequestID"=?`, testid).Update(statistics).Error; err != nil {
			return err
		}
		log.Debug("Successfully update data to the table Purch_Oppt from test_ID", testid)
		return nil
	}

	return nil
}

func (api netSenseAPI) insertCallsInfo(db *gorm.DB, tr testResult, lt foundTest) (string, error) {
	for i := range tr.Result.List {
		if i == 0 {
			testedFrom := tr.Result.List[i].CallResult.CallStart.Value
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

func (api netSenseAPI) checkPresentAudioFile(db *gorm.DB, tr testResult) {
	for _, l := range tr.Result.List {
		audioURL := l.CallResult.AudioURL.Value
		callID := l.CallResult.CallResultID.Value
		if audioURL == "" {
			log.Info("Not present audio file for call_id", callID)
			if err := insertEmptyFiles(db, callID); err != nil {
				log.Errorf(401, "Cann't update data row about empty request for call_id %s|%v", callID, err)
			}
			continue
		}
		log.Infof("Download AudioURL:%s for call_id:%s", audioURL, callID)
		request = fmt.Sprintf("%s/%s/%s", api.AudioFile, api.AuthKey, audioURL)
		// nameFile := strings.Split(audioURL, "/")[1]
		if err := saveWavFile(request, callID); err != nil {
			log.Errorf(402, "Cann't download the audio file for call_id:%s|%v", callID, err)
			continue
		}
		log.Infof("Succeseffuly download audio file for call_id:", callID)
		// cWav := body
		// ! Тут нужна функция вычисления координаты х, вертикальной черты ответа на звонок
		//! Пока она по умолчанию = 0
		cImg, err := waveFormImage(callID, 0)
		if err != nil || len(cImg) == 0 {
			log.Errorf(403, "Cann't create waveform image file for call_id %s|%v", callID, err)
			cImg = []byte("C&V:Cann't create waveform image file")
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
			DataLoaded:  true,
			AudioFile:   body,
			AudioGraph:  cImg,
			ConnectTime: l.CallResult.ConnectTime,
			CallType:    l.CallResult.CallType,
		}
		if err = updateCallsInfo(db, callID, callsinfo); err != nil {
			log.Errorf(404, "Cann't insert WAV file into table for system Assure call_id %s|%v", callID, err)
			continue
		}
		log.Info("Insert WAV and IMG file for callid", callID)

		if err = deleteFiles(listDeleteFiles); err != nil {
			log.Errorf(405, "Cann't delete some files for call_id %s|%v", callID, err)
		}
	}
}

func saveWavFile(request, nameFile string) error {
	res, err := requestGET(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("empty audio file")
	}
	if err := ioutil.WriteFile(srvTmpFolder+nameFile+".wav", body, 0666); err != nil {
		return err
	}
	return nil
}

// func (api netSenseAPI) destinationsync(db *gorm.DB) error {
// 	url := "https://netsense.arptel.com/netsense/" // из API
// 	req := "destinationsync/"                      // из API
// 	myurl := url + req
// 	AuthKey := "8e66ab96dd031d83308212f0567ae86b" // из API
// 	timeZone := "Europe/Stockholm"                // из settings?
// 	isLastRequest := "true"                       // из settings?
// 	header := fmt.Sprintf(`<destinationRequest>
// 	<authentication>
// 	  <key>%s</key>
// 	</authentication>
// 	<parameters>
// 	  <destinationList>`, AuthKey)
// 	var l []destinationList
// 	var body string
// 	for i := range l {
// 		d := fmt.Sprintf(`<list>
// 			<destination>%s</destination>
// 			<code>%s</code>
// 			<alias>%s</alias>
// 		  </list>`, l[i].destination, l[i].code, l[i].alias)
// 		body = body + d
// 	}
// 	footer := fmt.Sprintf(`</destinationList>
// 		</parameters>
// 		<settings>
// 		  <timeZone>%s</timeZone>
// 		  <isLastRequest>%s</isLastRequest>
// 		  <chainCode/>
// 		</settings>
// 	  </destinationRequest>`, timeZone, isLastRequest)

// 	xmlbody := header + body + footer
// 	resp, err := http.Post(myurl, "application/xml", strings.NewReader(xmlbody))
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	return nil

//! если их серевер при запросе будет требовать в заголовке поле Accept,
//! то запрос буду составлять по другому
//   or you can use []byte(`...`) and convert to Buffer later on
//   body := "<request> <parameters> <email>test@test.com</email> <password>test</password> </parameters> </request>"

//   client := &http.Client{}
// build a new request, but not doing the POST yet
//   req, err := http.NewRequest("POST", "http://localhost:8080/", bytes.NewBuffer([]byte(body)))
//   if err != nil {
// 	  fmt.Println(err)
//   }
// you can then set the Header here
// I think the content-type should be "application/xml" like json...
//   req.Header.Add("Content-Type", "application/xml; charset=utf-8")
// now POST it
//   resp, err := client.Do(req)
//   if err != nil {
// 	  fmt.Println(err)
//   }
//   fmt.Println(resp)
// }

// func (api netSenseAPI) routesync(db *gorm.DB) error {
// 	url := "https://netsense.arptel.com/netsense/" // из API
// 	req := "routesync/"
// 	myurl := url + req
// 	AuthKey := "8e66ab96dd031d83308212f0567ae86b"
// 	parentNode := " Parent Folder Path "                 //еще не знаю что это
// 	dialerGroup := " Dialer Group Path "                 //еще не знаю что это
// 	filter := "NO/GROUPED-ABC/DETAILED-ABC (Default NO)" //еще не знаю что это
// 	header := fmt.Sprintf(`<routeRequest>
// 	<authentication>
// 	  <key>%s</key>
// 	</authentication>
// 	<settings>
// 	  <parentNode>%s</parentNode>
// 	  <dialerGroup>%s</dialerGroup>
// 	  <filter>%s</filter>
// 	</settings>
// 	<parameters>
// 	  <routeList>`, AuthKey, parentNode, dialerGroup, filter)
// 	var l []routeList
// 	var body string
// 	for i := range l {
// 		d := fmt.Sprintf(`<list>
//         <externalId>%s</externalId>
//         <folder>%s</folder>
//         <carrier>%s</carrier>
//         <trunkGroup>%s</trunkGroup>
//         <trunk>%s</trunk>
//         <description>%s</description>
//         <dialerGroup>%s</dialerGroup>
//         <iac>%s</iac>
//         <nac>%s</nac>
//       </list>`, l[i].externalID, l[i].folder, l[i].carrier, l[i].trunkGroup, l[i].trunk, l[i].description, l[i].dialerGroup, l[i].iac, l[i].nac)
// 		body = body + d
// 	}
// 	footer := "</routeList></parameters></routeRequest>"

// 	xmlbody := header + body + footer

// 	resp, err := http.Post(myurl, "application/xml", strings.NewReader(xmlbody))
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	return nil
// }
