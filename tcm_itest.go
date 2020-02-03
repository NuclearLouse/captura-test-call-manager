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
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/net/html/charset"

	log "redits.oculeus.com/asorokin/my_packages/logging"
)

func (api *itestAPI) sysName(db *gorm.DB) string {
	db.Take(api)
	return api.SystemName
}

func (api itestAPI) checkAuth(db *gorm.DB) bool {
	return true
}

func (api itestAPI) runNewTest(db *gorm.DB, fnt foundTest) error {
	// тут нужен вызов промежуточной функции с преобразованиями значений Captura на
	// значения нужные Assure.
	// Пока будет просто пропуск.

	// var request string
	// если тест по Б-номеру
	//   https://api.i-­test.net/?t=2011&profid=12&prefix=34&numbers=1234%%-­5678%%-­12345678
	// switch fnt.TestType.name() {
	// case "cli":
	// 	request = fmt.Sprintf("%s?t=%d&profid=%s&vended=%s&ndbccgid=%s&ndbcgid=%s",
	// 		api.URL,
	// 		api.TestInitCli,
	// 		*.ProfileID,  //из CallingSys_iTest_profiles что это?
	// 		*.SupplierID, //из CallingSys_iTest_suppliers это как Route у Assure?
	// 		*.CountryID,  //из CallingSys_iTest_breakouts_cli это как Destination у Assure?
	// 		*.BreakoutID) //из CallingSys_iTest_breakouts_cli это как Destination у Assure?
	// case "voice":
	// 	request = fmt.Sprintf("%s?t=%d&profid=%s&prefix=%s&ndbccgid=%s&ndbcgid=%s",
	// 		api.URL,
	// 		api.TestInit,
	// 		*.ProfileID,  //из CallingSys_iTest_profiles
	// 		*.Prefix,     //из CallingSys_iTest_suppliers что это?
	// 		*.CountryID,  //из CallingSys_iTest_breakouts_std
	// 		*.BreakoutID) //из CallingSys_iTest_breakouts_std
	// }
	// response, err := api.requestPOST(request)
	// if err != nil {
	// 	return err
	// }
	// decoder := xmlDecoder(response)
	// var testinit TestInitiation
	// if err := decoder.Decode(&testinit); err != nil {
	// 	return err
	// }

	// newTestInfo := PurchOppt{
	// 	TestingSystemRequestID: testinit.Test.TestID,
	// 	TestComment:            testinit.Test.ShareURL,
	// 	RequestState:           2}
	// if err := db.Model(&newTestInfo).Where(`"RequestID"=?`, ft.RequestID).Update(newTestInfo).Error; err != nil {
	// 	return err
	// }
	return nil
}

func (api itestAPI) requestPOST(req string) (*http.Response, error) {
	res, err := http.PostForm(req, url.Values{"email": {api.User}, "pass": {api.Pass}})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (api itestAPI) checkPresentAudioFile(call, callID string) (string, bool, error) {
	var req, nameFile string
	pref := fmt.Sprintf("%v_", time.Now().Unix())
	switch call {
	case "beep":
		nameFile = pref + callID + "-r.mp3"
		req = fmt.Sprintf("%s%s/call-%s-r.mp3", api.RepoURL, callID[:8], callID)
	case "answer":
		nameFile = pref + callID + ".mp3"
		req = fmt.Sprintf("%s%s/call-%s.mp3", api.RepoURL, callID[:8], callID)
	}
	res, err := api.requestPOST(req)
	if err != nil {
		return "error", false, err
	}
	if res.Header[`Content-Type`][0] != "audio/mpeg" {
		// errstring := fmt.Sprintf("Test system didn't provide audio file, response return Content-Type=%s", response.Header[`Content-Type`][0])
		return "html", false, errors.New("Not present audio file")
	}
	file, err := createFile(res, nameFile)
	if err != nil {
		return "error", false, err
	}
	name := strings.Split(nameFile, ".")[0]
	return name, file, nil
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
	log.Debugf("Successful response to the request Complete_Test for system %s test_ID %s", sysname, testid)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()
	defer func() {
		if r := recover(); r != nil {
			log.Info("Wait. The test has not yet begun for test_ID", testid)
		}
	}()
	//TODO: Здесь надо получать и аудио УРЛ и вставлять его в таблицу
	re1 := regexp.MustCompile(`<Calls_Total>[0-9]+`)
	fs1 := re1.FindAllString(string(body), -1)
	ss1 := strings.Split(fs1[0], "<Calls_Total>")
	re2 := regexp.MustCompile(`<Calls_Complete>[0-9]+`)
	fs2 := re2.FindAllString(string(body), -1)
	ss2 := strings.Split(fs2[0], "<Calls_Complete>")
	if ss1[1] == ss2[1] {
		req := fmt.Sprintf("%s?t=%d&jid=%s", api.URL, api.TestStatusDetails, testid)
		res, err := api.requestPOST(req)
		log.Debugf("Sending request TestResults fot system %s test_ID %s", sysname, testid)
		if err != nil {
			return err
		}
		log.Infof("Successful response to the request TestResults for system %s test_ID %s", sysname, testid)
		decoder := xmlDecoder(res)
		var callsinfo CallsInfo
		if err := decoder.Decode(&callsinfo); err != nil {
			return err
		}
		start := time.Now()
		log.Debugf("Start transaction insert into the table TestResults for system %s test_id %s", sysname, testid)
		if err := api.insertCallsInfo(db, callsinfo, lt); err != nil {
			return err
		}
		log.Infof("Successfully insert data from table TestResults for system %s test_ID %s", sysname, testid)
		log.Debug("Elapsed time insert transaction", time.Since(start))
		var statistics PurchOppt
		statistics = callsStatistics(db, testid)
		statistics.TestedFrom = time.Unix(callsinfo.TestOverview.Init, 0)
		if err = db.Model(&statistics).Where(`"TestingSystemRequestID"=?`, testid).Update(statistics).Error; err != nil {
			return err
		}
		log.Info("Successfully update data to the table Purch_Oppt from test_ID", testid)
		return nil
	}
	log.Info("Wait. The test is not over yet for test_ID", testid)
	return nil
}

func (api itestAPI) uploadResultFiles(db *gorm.DB) {
	for {
		var rows []CallingSysTestResults
		if err := db.Where(`"DataLoaded"=false AND "TestSystem"=?`, api.SystemID).Find(&rows).Error; err != nil {
			log.Errorf(502, "Cann't obtain rows for DataLoaded=false|%v", err)
			continue
		}
		if len(rows) == 0 {
			// This log will be every second in the absence of new tests
			log.Trace("Not rows for files_uploaded=false. All files upload.")
			time.Sleep(time.Duration(1) * time.Second) // just in case
			continue
		}
		for i := range rows {
			var fileBeep, fileAnsw bool
			var nameFile, nameFileBeep, nameFileAnsw string
			var err error
			var cWav []byte
			var cImg []byte
			if rows[i].RingDuration == -1 && rows[i].CallDuration == -1 {
				log.Debug("Not present beep and answer audio file for call_id", rows[i].CallID)
				if err = insertEmptyFiles(db, rows[i].CallID); err != nil {
					log.Errorf(503, "Cann't update data row about empty request for call_id %s|%v", rows[i].CallID, err)
				}
				continue
			}
			if rows[i].CallDuration == -1 {
				nameFileBeep, fileBeep, err = api.checkPresentAudioFile("beep", rows[i].CallID)
				switch nameFileBeep {
				case "html":
					log.Info("Not present audio file for call_id", rows[i].CallID)
					if err = insertEmptyFiles(db, rows[i].CallID); err != nil {
						log.Errorf(503, "Cann't update data row about empty request for call_id %s||%v", rows[i].CallID, err)
					}
					continue
				case "error":
					log.Errorf(504, "For call_id %s|%v", rows[i].CallID, err)
					continue
				}
				if fileBeep {
					log.Info("Download mp3 beep file for call_id", rows[i].CallID)
				}
				cWav, err = decodeToWAV(nameFileBeep, "mp3")
				if err != nil {
					log.Errorf(505, "Cann't decode MP3 file to WAV for call_id %s|%v", rows[i].CallID, err)
					continue
				}
				log.Info("Decode mp3 file to WAV for call_id", rows[i].CallID)
				cImg, err = waveFormImage(nameFileBeep, 0)
				if err != nil || len(cImg) == 0 {
					log.Errorf(506, "Cann't create waveform PNG image file for call_id %s|%v", rows[i].CallID, err)
					cImg = labelEmptyBMP("C&V:Cann't create waveform image file")
					continue
				}
				log.Info("Created image PNG file for call_id", rows[i].CallID)

				listDeleteFiles := []string{
					srvTmpFolder + nameFileBeep + ".mp3",
					srvTmpFolder + nameFileBeep + ".wav",
					srvTmpFolder + nameFileBeep + ".png",
					srvTmpFolder + nameFileBeep + ".bmp",
				}

				callsinfo := CallingSysTestResults{
					DataLoaded:  true,
					AudioFile:   cWav,
					AudioGraph:  cImg,
					ConnectTime: 0,
				}
				if err = updateCallsInfo(db, rows[i].CallID, callsinfo); err != nil {
					log.Errorf(507, "Cann't insert WAV file into table for system %s call_id %s|%v", api.SystemName, rows[i].CallID, err)
					continue
				}
				log.Info("Insert WAV and IMG file for callid", rows[i].CallID)

				if err = deleteFiles(listDeleteFiles); err != nil {
					log.Errorf(508, "Cann't delete beep or answer mp3 files for call_ID %s|%v", rows[i].CallID, err)
				}
				continue
			}
			fileBeep = false
			fileAnsw = false
			if rows[i].CallDuration != -1 {
				nameFileBeep, fileBeep, err = api.checkPresentAudioFile("beep", rows[i].CallID)
				switch nameFileBeep {
				case "html":
					log.Warnf("%v for call_id %s", err, rows[i].CallID)
				case "error":
					log.Errorf(504, "For call_id %s|%v", rows[i].CallID, err)
					continue
				}
				if fileBeep {
					log.Info("Download mp3 beep file for call_id", rows[i].CallID)
				}
				nameFileAnsw, fileAnsw, err = api.checkPresentAudioFile("answer", rows[i].CallID)
				switch nameFileAnsw {
				case "html":
					log.Warnf("%v for call_id %s", err, rows[i].CallID)
				case "error":
					log.Errorf(504, "For call_id %s|%v", rows[i].CallID, err)
					continue
				}
				if fileAnsw {
					log.Info("Download mp3 answ file for call_id", rows[i].CallID)
				}
				var connectTime float64
				var x int
				switch {
				case fileBeep && fileAnsw:
					connectTime, x, err = calcCoordinate(nameFileBeep, nameFileAnsw)
					if err != nil {
						log.Errorf(509, "Cann't get the coordinate of the beginning of the answer for call_ID %s|%v", rows[i].CallID, err)
					}
					nameFile, err = concatMP3files(nameFileBeep, nameFileAnsw)
					if err != nil {
						log.Errorf(510, "Cann't concatenate beep and answer MP3 files for call_id %s|%v", rows[i].CallID, err)
						continue
					}
					log.Info("Concatenate beep and answer mp3 files for call_id", rows[i].CallID)
					listDeleteFiles := []string{
						srvTmpFolder + nameFileBeep + ".mp3",
						srvTmpFolder + nameFileAnsw + ".mp3",
					}
					if err = deleteFiles(listDeleteFiles); err != nil {
						log.Errorf(508, "Cann't delete beep or answer mp3 files for call_id %s|%v", rows[i].CallID, err)
					}
				case fileBeep && !fileAnsw:
					nameFile = nameFileBeep
				case !fileBeep && fileAnsw:
					nameFile = nameFileAnsw
				case !fileBeep && !fileAnsw:
					log.Debug("There are not beep and answer audio files for call_id", rows[i].CallID)
					if err = insertEmptyFiles(db, rows[i].CallID); err != nil {
						log.Errorf(503, "Cann't update data row about empty request for call_id %s|%v", rows[i].CallID, err)
					}
					continue
				}
				cWav, err = decodeToWAV(nameFile, "mp3")
				if err != nil {
					log.Errorf(505, "Cann't decode MP3 file to WAV for call_id %s|%v", rows[i].CallID, err)
					continue
				}
				log.Info("Decod mp3 files to WAV for call_id", rows[i].CallID)
				cImg, err = waveFormImage(nameFile, x)
				if err != nil || len(cImg) == 0 {
					log.Errorf(506, "Cann't create waveform PNG image file for call_id %s|%v", rows[i].CallID, err)
					cImg = labelEmptyBMP("C&V:Cann't create waveform image file")
					continue
				}
				log.Info("Created image PNG file for call_id", rows[i].CallID)

				listDeleteFiles := []string{
					srvTmpFolder + nameFile + ".mp3",
					srvTmpFolder + nameFile + ".wav",
					srvTmpFolder + nameFile + ".png",
					srvTmpFolder + nameFile + ".bmp",
				}

				callsinfo := CallingSysTestResults{
					DataLoaded:  true,
					AudioFile:   cWav,
					AudioGraph:  cImg,
					ConnectTime: connectTime,
				}
				if err = updateCallsInfo(db, rows[i].CallID, callsinfo); err != nil {
					log.Errorf(507, "Cann't insert WAV file into table for call_id %s|%v", rows[i].CallID, err)
					continue
				}
				log.Info("Insert WAV and IMG file for callid", rows[i].CallID)

				if err = deleteFiles(listDeleteFiles); err != nil {
					log.Errorf(512, "Cann't delete beep or answer mp3 files for call_id %s|%v", rows[i].CallID, err)
				}

			}
		}
		log.Infof("All present files download from %s server and upload into the table TestResults", api.SystemName)
		time.Sleep(time.Duration(1) * time.Second) // just in case
	}

}

func (itestAPI) insertCallsInfo(db *gorm.DB, ci CallsInfo, ti foundTest) error {
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
	var ringDuration, callDuration float64
	for _, call := range ci.Calls {
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
			CallID:                   call.CallID,
			CallListID:               ci.TestOverview.TestID, //ti.TestingSystemRequestID,
			TestSystem:               ti.SystemID,
			CallType:                 ci.TestOverview.Type,
			Destination:              call.Destination,
			CallStart:                time.Unix(call.Start, 0),
			CallComplete:             time.Unix(int64(call.End), 0),
			CallDuration:             callDuration,
			RingDuration:             ringDuration,
			PDD:                      call.PDD,
			BNumber:                  prefixAndBnumber[1],
			BNumberDialed:            call.Destination,
			CallingNumber:            call.Source,
			Route:                    ci.TestOverview.Supplier, //ti.RouteCarrier,
			CauseCodeID:              call.ResultCode,
			Status:                   call.LastCode,
			CliDetectedCallingNumber: call.CLI,
			CliResult:                call.Result,
			FasResult:                call.FAS,
			VoiceQualityMos:          call.MOS,
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

func xmlDecoder(res *http.Response) *xml.Decoder {
	decoder := xml.NewDecoder(res.Body)
	decoder.Strict = false
	decoder.CharsetReader = charset.NewReaderLabel
	return decoder
}

func createFile(res *http.Response, nameFile string) (bool, error) {
	filepath := srvTmpFolder + nameFile
	file, err := os.Create(filepath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return false, err
	}
	_, err = ioutil.ReadFile(filepath)
	if err != nil {
		return false, err
	}
	return true, nil
}

func concatMP3files(fileBeep, fileAnsw string) (string, error) {
	pathBeep := srvTmpFolder + fileBeep + ".mp3"
	pathAnsw := srvTmpFolder + fileAnsw + ".mp3"
	pathOut := srvTmpFolder + "out_" + fileAnsw + ".mp3"
	com := fmt.Sprintln(fmt.Sprintf(ffmpegConcatMP3, pathBeep, pathAnsw, pathOut))
	_, err := execCommand(com)
	if err != nil {
		return "", err
	}
	return "out_" + fileAnsw, nil
}

func calcCoordinate(fileBeep, fileAnsw string) (float64, int, error) {
	var files [2]string
	files[0] = srvTmpFolder + fileBeep + ".mp3"
	files[1] = srvTmpFolder + fileAnsw + ".mp3"
	var duration [2]int
	for i := 0; i < len(files); i++ {
		com := fmt.Sprintln(fmt.Sprintf(ffmpegDuration, files[i]))
		out, err := execCommand(com)
		if err != nil {
			return 0, 0, err
		}
		strOut := strings.Split(string(out), ",")
		strTime := strings.Split(strOut[0], "Duration:")
		t := iTestParseTime(strings.TrimSpace(strTime[1]))
		duration[i] = 60*t.Minute() + t.Second()
	}
	return float64(duration[0]), 500 * duration[0] / (duration[0] + duration[1]), nil
}
