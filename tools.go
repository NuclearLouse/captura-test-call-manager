// assistants.go
//
// The file contains auxiliary functions that are necessary for the operation of all test systems.
//
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fogleman/gg"

	"github.com/jinzhu/gorm"
	log "redits.oculeus.com/asorokin/my_packages/logging"
)

// var message = fmt.Sprintf

func createTables(db *gorm.DB) error {
	listTables := []interface{}{
		&itestBreakouts{BreakType: "std"},
		&itestBreakouts{BreakType: "cli"},
		&itestProfiles{},
		&itestSuppliers{},
		&itestAPI{},
		&netSenseAPI{},
		&assureAPI{},
		&assureRoutes{},
		&assureDestinations{},
		&assureNodes{},
		&assureNodesCapabilities{},
	}
	var errs error
	for i := range listTables {
		// if err := db.AutoMigrate(listTables[i]).Error; err != nil {
		// 	log.Errorf(9, "Cann't create table|%v", err)
		// 	errs = err
		// }
		if !db.HasTable(listTables[i]) {
			if err := db.CreateTable(listTables[i]).Error; err != nil {
				if !strings.Contains(err.Error(), "already exists") {
					log.Errorf(9, "Cann't create table|%v", err)
					errs = err
				}
			}
		}
	}
	return errs
}

func createDir(nameDir string) error {
	_, err := os.Stat(nameDir)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(nameDir, 0777); err != nil {
			return err
		}
	}
	return nil
}

func deleteFiles(nameFiles []string) error {
	var err error
	for i := range nameFiles {
		if err = os.Remove(nameFiles[i]); err == nil {
			log.Debug("Successefuly delete file", nameFiles[i])
		}
	}
	return err
}

func execCommand(com string) ([]byte, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("bash", "-c", com)
	case "windows":
		cmd = exec.Command("cmd", "/C", com)
	}
	out, err := cmd.Output()
	if err != nil {
		return out, err
	}
	cmd.Wait()
	return out, nil
}

func waveFormImage(nameFile string, x float64) ([]byte, error) {
	pathWavFile := os.Getenv("ABS_PATH_DWL") + nameFile + ".wav"
	pathPngFile := os.Getenv("ABS_PATH_DWL") + nameFile + ".png"
	com := fmt.Sprintf("%s -i %s -lavfi showwavespic=split_channels=1:s=600x400:colors=000000 %s",
		os.Getenv("FFMPEG"), pathWavFile, pathPngFile)
	_, err := execCommand(com)
	if err != nil {
		return nil, err
	}
	img, err := gg.LoadPNG(pathPngFile)
	if err != nil {
		log.Errorf(4, "Could not draw a horizontal line on the wave image for file %s.png|%v", nameFile, err)
	}
	dl := gg.NewContext(600, 400)
	dl.Clear()
	dl.SetRGB255(0, 0, 0)
	dl.DrawRectangle(0, 199, 600, 2)
	dl.Fill()
	dl.DrawImage(img, 0, 0)
	dl.SavePNG(pathPngFile)

	if strings.HasPrefix(nameFile, "out_") {
		if err := drawVLine(pathPngFile, x); err != nil {
			log.Errorf(5, "Could not draw a vertical line on the wave image for file %s.png|%v", nameFile, err)
		}
		log.Debugf("Successefuly draw vertical line for file %s.wav", nameFile)
	}
	content, err := ioutil.ReadFile(pathPngFile)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func drawVLine(pathPngFile string, x float64) error {
	img, err := gg.LoadPNG(pathPngFile)
	if err != nil {
		return err
	}
	dl := gg.NewContext(600, 400)
	dl.Clear()

	dl.SetRGB255(255, 0, 0)
	dl.DrawRectangle(x, 0, 4, 400)
	dl.Fill()

	dl.DrawImage(img, 0, 0)
	dl.SavePNG(pathPngFile)
	return nil
}

func decodeToWAV(nameFile, codec string) ([]byte, error) {
	file := os.Getenv("ABS_PATH_DWL") + nameFile + "." + codec
	fileWAV := os.Getenv("ABS_PATH_DWL") + nameFile + ".wav"
	com := fmt.Sprintf("%s -y -i %s %s", os.Getenv("FFMPEG"), file, fileWAV)
	_, err := execCommand(com)
	// log.Debug(com)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(fileWAV)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return content, nil
}

func deleteOldTestInfo(db *gorm.DB) error {
	query := fmt.Sprintf(`DELETE FROM %s"CallingSys_TestResults" AS t1 
	USING %[1]s"CallingSys_Settings" AS t2 
	WHERE t1."TestSystem"=t2."SystemID" 
	AND (CURRENT_TIMESTAMP::date-t1."CallComplete"::date)>t2."Log_Period";`, os.Getenv("SCHEMA_PG"))
	return db.Exec(query).Error
}

func isEnabled(db *gorm.DB, enb bool, name string) (CallingSysSettings, error) {
	var sys CallingSysSettings
	if err := db.Where(`"Enabled"=? AND "SystemName"=?`, enb, name).Find(&sys).Error; err != nil {
		return sys, err
	}
	return sys, nil
}

func updateAPI(db *gorm.DB, model interface{}, sys CallingSysSettings) *gorm.DB {
	return db.Model(model).Updates(map[string]interface{}{"url": sys.Address, "user": sys.AuthName, "pass": sys.AuthKey}).Take(model)
}

// TODO: возможно потом сюда лучше будет передавать уже подготовленную структуру, а не набор ее полей
func updateCallsInfo(db *gorm.DB, callID string, callsinfo CallingSysTestResults) error {
	if err := db.Model(&callsinfo).Where(`"CallID"=?`, callID).Updates(callsinfo).Error; err != nil {
		return err
	}
	return nil
}

func insertEmptyFiles(db *gorm.DB, callID string) error {
	callsinfo := CallingSysTestResults{
		DataLoaded:  true,
		AudioFile:   []byte("C&V:test system didn't provide audio files"),
		AudioGraph:  []byte("C&V:test system didn't provide audio files"),
		ConnectTime: 0,
	}
	if err := updateCallsInfo(db, callID, callsinfo); err != nil {
		return err
	}
	return nil
}

// CallsStatistics returns a structure for adding statistics to the PurchOppt table
// For each test by its testID, I find the number of calls [Count(&total)],
// find the end time of the last call [MAX("CallComplete")]
// Counting the number of calls with a duration> 0,
// and summarize the total duration of these calls
func callsStatistics(db *gorm.DB, testid string) PurchOppt {
	var total, complete, sumcalls float64
	var max time.Time
	var testresult CallingSysTestResults
	db.Model(&testresult).
		Where(`"CallListID" = ?`, testid).
		Select(`MAX("CallComplete")`).
		Count(&total).
		Row().
		Scan(&max)
	db.Model(&testresult).
		Where(`"CallListID" = ? AND "CallDuration" > 0`, testid).
		Select(`SUM("CallDuration")`).
		Count(&complete).
		Row().
		Scan(&sumcalls)
	stat := PurchOppt{
		RequestState: 2,
		TestedUntil:  max,
		TestASR:      100 * complete / total,
		TestACD:      sumcalls / complete / 60,
		// TestCalls:    int(total),
		TestMinutes: sumcalls / 60,
	}
	return stat
}
