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

	log "captura_tcm/logger"

	"github.com/jinzhu/gorm"
)

func (api *netSenseAPI) sysName(db *gorm.DB) string {
	db.Take(api)
	return api.SystemName
}

func (api *netSenseAPI) sysID(db *gorm.DB) int {
	db.Take(api)
	return api.SystemID
}

func (api netSenseAPI) checkAuth(db *gorm.DB) bool {
	return true
}

func (api netSenseAPI) prepareRequests(db *gorm.DB, interval int64) {
	log.Info("Send preparatory requests for", api.SystemName)
	log.Debug("API Settings", api)
	for {
		log.Infof("The next data update to prepare %s after %d hours", api.SystemName, interval)
		time.Sleep(time.Duration(interval) * time.Hour)
	}

}

func (api netSenseAPI) newRequest(method, request string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, api.URL+request, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/xml")
	}

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

func (api netSenseAPI) cancelTest(db *gorm.DB, testid string) error {
	log.Debugf("Sending a request Cancel Test for system %s and test_id %s", api.SystemName, testid)

	return nil
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
	res, err := api.newRequest("POST", api.TestInitList, xmlBody)
	if err != nil {
		return err
	}

	var newTests testInitResponse
	if err := xml.NewDecoder(res.Body).Decode(&newTests); err != nil {
		return err
	}
	res.Body.Close()

	switch newTests.CallListResponseArray.Status.Code {
	case "200":
		testinfo := purchOppt{
			TestingSystemRequestID: newTests.CallListResponseArray.ResponseID,
			RequestState:           2}
		if err := testinfo.updateTestInfo(db, nit.RequestID); err != nil {
			return err
		}
	case "400":
		message := fmt.Sprintf("Bad Request.Error code:%s.%s\n", newTests.CallListResponseArray.SubStatus.List.Status.Code, newTests.CallListResponseArray.SubStatus.List.Status.Message)
		return errors.New(message)
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

	res, err := api.newRequest("POST", api.TestInitList, xmlBody)
	if err != nil {
		return err
	}

	var ts testStatusNetsense
	if err := xml.NewDecoder(res.Body).Decode(&ts); err != nil {
		return err
	}
	res.Body.Close()

	var ti purchOppt
	ti.TestResult = ts.CallListResponseArray.CallLogResponses.Status
	switch ti.TestResult {
	case "RUNNING":
		log.Debug("Wait. The test is not over yet for test_ID", testid)
	case "END":
		log.Info("The end test for test_ID", testid)
		request := fmt.Sprintf("%s/%s/1/%s", api.TestInit, api.AuthKey, ts.CallListResponseArray.CallLogResponses.CallListLogID)
		res, err := api.newRequest("GET", request, nil)
		log.Debugf("Sending request TestResults for system Assure test_ID %s", testid)
		if err != nil {
			return err
		}

		log.Infof("Successful response to the request TestResults for system Assure test_ID %s", testid)

		var callsinfo testResultNetsense
		if err := xml.NewDecoder(res.Body).Decode(&callsinfo); err != nil {
			return err
		}
		res.Body.Close()

		start := time.Now()
		log.Debugf("Start transaction insert into the table TestResults for system Netsense test_id %s", testid)
		testedFrom, err := api.insertCallsInfo(db, callsinfo, lt)
		if err != nil {
			return err
		}
		log.Infof("Successfully insert data from table TestResults for system Netsense test_ID %s", testid)
		log.Debug("Elapsed time insert transaction", time.Since(start))

		ti.callsStatistic(db, testid)
		ti.TestedFrom = netsenseParseTime(testedFrom)
		ti.TestedByUser = lt.RequestByUser
		// statistic.TestComment = "Bla bla bla test by Assure for test_ID:" + testid
		go api.downloadAudioFiles(db, callsinfo)

	default:
		// ?Most likely this situation will never arise
		log.Info("Failed test for test_ID", testid)
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

func (api netSenseAPI) insertCallsInfo(db *gorm.DB, tr testResultNetsense, lt foundTest) (string, error) {
	var testedFrom string
	for i := range tr.Result.List {
		if i == 0 {
			testedFrom = tr.Result.List[i].CallResult.CallStart.Value
		}
		res := tr.Result.List[i].CallResult
		resCLI := tr.Result.List[i].CallResultCLI
		resFAS := tr.Result.List[i].CallResultFAS
		callinfo := callingSysTestResults{
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

func (api netSenseAPI) downloadAudioFiles(db *gorm.DB, tr testResultNetsense) {
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
		if err != nil || len(cWav) == 0 {
			log.Errorf(402, "Cann't download the audio file for call_id:%s|%v", callID, err)
			continue
		}
		log.Info("Succeseffuly download wav file for call_id:", callID)
		cMP3, err := decodeAudio(callID, "wav", "mp3", "result")
		if err != nil || len(cMP3) == 0 {
			log.Errorf(400, "Cann't decode MP3 file for call_id %s|%v", callID, err)
		}
		var x int
		if l.CallResult.CallType.Value != "TrueCLI" {
			// TODO: Формулу потом надо пересмотреть
			x, _ = strconv.Atoi(fmt.Sprintf("%.f", 500*l.CallResult.ConnectTime.Value/(l.CallResult.ConnectTime.Value+l.CallResult.CallDuration.Value)))
		}

		pngImg, bmpImpg, err := waveFormImage(callID, x)
		if err != nil || len(bmpImpg) == 0 || len(pngImg) == 0 {
			log.Errorf(403, "Cann't create waveform image file for call_id %s|%v", callID, err)
			bmpImpg = labelEmptyBMP("C&V:Cann't create waveform image file")
			continue
		}
		log.Info("Created image files for call_id", callID)

		callsinfo := callingSysTestResults{
			DataLoaded: true,
			AudioFile:  cWav,
			AudioGraph: bmpImpg,
		}
		if err = callsinfo.updateCallsInfo(db, callID); err != nil {
			log.Errorf(404, "Cann't insert WAV file into TestResults table for call_id %s|%v", callID, err)
			continue
		}

		webinfo := testFilesWEB{
			Callid:     callID,
			Testsystem: api.SystemID,
			Diagram:    pngImg,
			Audiofile:  cMP3}
		if err = webinfo.insertWebInfo(db); err != nil {
			log.Errorf(405, "Cann't insert audio and image file into testfiles_web table for call_id %s|%v", callID, err)
			continue
		}

		log.Info("Insert WAV and IMG file for callid", callID)
	}
}

func (api netSenseAPI) saveWavFile(request, nameFile string) ([]byte, error) {
	res, err := api.newRequest("GET", request, nil)
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

func (netSenseAPI) TableName() string {
	return schemaPG + "CallingSys_API_NetSense"
}

type netSenseAPI struct {
	SystemName       string `gorm:"size:50"`
	SystemID         int    `gorm:"type:int"`
	URL              string `gorm:"size:100"`
	User             string `gorm:"size:100"`
	AuthKey          string `gorm:"column:pass;size:100"`
	TestInit         string `gorm:"size:100"`
	TestInitList     string `gorm:"size:100"`
	GetCallListid    string `gorm:"size:100"`
	GetDestinations  string `gorm:"size:50"`
	SyncDestinations string `gorm:"size:50"`
	SyncRoutes       string `gorm:"size:50"`
	AudioFile        string `gorm:"size:50"`
	GetAllCalls      string `gorm:"size:50"`
	Version          string `gorm:"size:50"`
}

//-----------------------------------------------------------------------------
//*******************Block of test initialization structures*******************
//-----------------------------------------------------------------------------
type testInitNetsense struct {
	XMLName        xml.Name `xml:"callListRequest"`
	Authentication authentication
	ParametersList parametersList
}

type authentication struct {
	XMLName xml.Name `xml:"authentication"`
	Key     string   `xml:"key"`
}

type parametersList struct {
	XMLName         xml.Name `xml:"parametersList"`
	CallTypeList    callTypeList
	RouteList       routeList
	DestinationList destinationList
}

type callTypeList struct {
	XMLName xml.Name `xml:"callTypeList"`
	List    listCallType
}

type listCallType struct {
	XMLName  xml.Name `xml:"list"`
	CallType string   `xml:"callType"`
}

type routeList struct {
	XMLName   xml.Name `xml:"routeList"`
	ListRoute listRoute
}

type listRoute struct {
	XMLName xml.Name `xml:"list"`
	Route   string   `xml:"route"`
}

type destinationList struct {
	XMLName xml.Name `xml:"destinationList"`
	List    []listDestination
}

type listDestination struct {
	XMLName     xml.Name `xml:"list"`
	Destination string   `xml:"destination"`
}

type testInitResponse struct {
	XMLName               xml.Name `xml:"callListResponseArray"`
	CallListResponseArray struct {
		Status struct {
			Code    string `xml:"code"`
			Message string `xml:"message"`
		} `xml:"status"`
		SubStatus struct {
			List struct {
				Parameters struct {
					CallTypeList struct {
						List struct {
							CallType string `xml:"callType"`
						} `xml:"list"`
					} `xml:"callTypeList"`
					DestinationList struct {
						List struct {
							Destination string `xml:"destination"`
						} `xml:"list"`
					} `xml:"destinationList"`
				} `xml:"parameters"`
				Status struct {
					Code    string `xml:"code"`
					Message string `xml:"message"`
				} `xml:"status"`
			} `xml:"list"`
		} `xml:"subStatus"`
		ResponseID string `xml:"responseId"`
	} `xml:"callListResponseArray"`
}

//-----------------------------------------------------------------------------
//**************Block of test status and test results structures***************
//-----------------------------------------------------------------------------
type testStatusNetsense struct {
	XMLName               xml.Name `xml:"callListResponseIdArray"`
	CallListResponseArray struct {
		CallListID       string `xml:"callListId"`
		CallLogResponses struct {
			CallListLogID string `xml:"callListLogId"`
			Status        string `xml:"status"`
		} `xml:"callLogResponses"`
	} `xml:"callListResponseArray"`
}

type getStatus struct {
	XMLName  xml.Name `xml:"callListIdRequest"`
	CallList string   `xml:"callListIdList"`
}

type testResultNetsense struct {
	XMLName xml.Name `xml:"response"`
	Result  struct {
		List []struct {
			CallResult struct {
				AlertTime struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"alertTime"`
				AudioURL struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"audioUrl"`
				CallComplete struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callComplete"`
				CallCompleteMs struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callCompleteMs"`
				CallDuration struct {
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"callDuration"`
				CallListLogID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callListLogId"`
				CallListRepeatNr struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callListRepeatNr"`
				CallListRowNr struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callListRowNr"`
				CallResultID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callResultId"`
				CallStart struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callStart"`
				CallStartMs struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callStartMs"`
				CallType struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callType"`
				CallingNumber struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callingNumber"`
				CauseCodeID struct {
					Format string `xml:"format"`
					Value  int    `xml:"value"`
				} `xml:"causeCodeId"`
				CauseLocationID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"causeLocationId"`
				ConnectTime struct {
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"connectTime"`
				Country struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"country"`
				Destination struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"destination"`
				DialedPhoneNumber struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dialedPhoneNumber"`
				Dialer struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dialer"`
				DialerGroup struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dialerGroup"`
				DisconnectTime struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"disconnectTime"`
				DtmfDetected struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dtmfDetected"`
				ExternalBatchID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"externalBatchId"`
				ExternalCallID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"externalCallId"`
				Label struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"label"`
				LastDigitSent struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"lastDigitSent"`
				MatchPercentage struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"matchPercentage"`
				MediaStartTime struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"mediaStartTime"`
				OntheFlyPhoneNumberName struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"ontheFlyPhoneNumberName"`
				Pdd struct {
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"pdd"`
				PddAlternate struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"pddAlternate"`
				PhoneNumber struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"phoneNumber"`
				PhoneNumberName struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"phoneNumberName"`
				PhoneNumberType struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"phoneNumberType"`
				Region struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"region"`
				RingingDuration struct {
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"ringingDuration"`
				Route struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"route"`
				RouteFolder struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"routeFolder"`
				SipCallID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"sipCallId"`
				Status struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"status"`
				StatusCode struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"statusCode"`
				ToneDetection struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"toneDetection"`
				Username struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"username"`
				WebServiceID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"webServiceId"`
			} `xml:"callResult"`
			CallResultCLI struct {
				CLIDetectedCallingNumber struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIDetectedCallingNumber"`
				CLIResult struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIResult"`
				CLIStatus struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIStatus"`
				CLIStatusCode struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIStatusCode"`
				CLIType struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIType"`
				CallResultID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callResultId"`
			} `xml:"callResultCLI"`
			CallResultFAS struct {
				CallResultID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callResultId"`
				FasResult struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"fasResult"`
				FasStatus struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"fasStatus"`
				FasStatusCode struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"fasStatusCode"`
			} `xml:"callResultFAS"`
		} `xml:"list"`
	} `xml:"result"`
	Status struct {
		Code    string `xml:"code"`
		Message string `xml:"message"`
	} `xml:"status"`
}

