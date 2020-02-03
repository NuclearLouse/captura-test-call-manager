// assistants.go
//
// The file contains auxiliary functions that are necessary for the operation of all test systems.
//
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/image/bmp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"golang.org/x/net/html/charset"
	log "redits.oculeus.com/asorokin/my_packages/logging"
)

// var message = fmt.Sprintf

func createTables(db *gorm.DB) error {
	listTables := []interface{}{
		&itestAPI{},
		&netSenseAPI{},
		&assureAPI{},
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
	case "windows":
		cmd = exec.Command("cmd", "/C", com)
	default:
		cmd = exec.Command("bash", "-c", com)
	}
	out, err := cmd.Output()
	if err != nil {
		return out, err
	}
	cmd.Wait()
	return out, nil
}

func waveFormImage(nameFile string, x int) ([]byte, error) {
	pathWavFile := srvTmpFolder + nameFile + ".wav"
	pathPngFile := srvTmpFolder + nameFile + ".png"
	com := fmt.Sprintln(fmt.Sprintf(ffmpegWavFormImg, pathWavFile, pathPngFile))
	_, err := execCommand(com)
	if err != nil {
		return nil, err
	}

	// drawing a vertical red line indicating the beginning of the answer
	// if strings.HasPrefix(nameFile, "out_") {
	if x != 0 {
		if err := drawVLine(pathPngFile, x); err != nil {
			return nil, err
		}
	}

	pathBmpFile := srvTmpFolder + nameFile + ".bmp"
	if err := encodePNGtoBMP(pathPngFile, pathBmpFile); err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(pathBmpFile)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func drawVLine(pathPngFile string, x int) error {
	filePNG, err := os.Open(pathPngFile)
	if err != nil {
		return err
	}
	img, err := png.Decode(filePNG)
	if err != nil {
		return err
	}
	filePNG.Close()
	dstImg := image.NewRGBA(img.Bounds())
	draw.Draw(dstImg, img.Bounds(), img, image.ZP, draw.Src)

	vline := image.Rect(x, 0, x+2, 100)
	draw.Draw(dstImg, vline, &image.Uniform{color.RGBA{255, 0, 0, 255}}, image.ZP, draw.Src)

	newImg, err := os.Create(pathPngFile)
	if err != nil {
		return err
	}
	defer newImg.Close()
	if err := png.Encode(newImg, dstImg); err != nil {
		return err
	}
	return nil
}

func encodePNGtoBMP(pathPngFile, pathBmpFile string) error {
	filePNG, err := os.Open(pathPngFile)
	if err != nil {
		return err
	}
	defer filePNG.Close()

	img, _, err := image.Decode(filePNG)
	if err != nil {
		return err
	}

	fileBMP, err := os.Create(pathBmpFile)
	if err != nil {
		return err
	}
	defer fileBMP.Close()

	if err := bmp.Encode(fileBMP, img); err != nil {
		return err
	}
	return nil
}

func decodeToWAV(nameFile, codec string) ([]byte, error) {
	file := srvTmpFolder + nameFile + "." + codec
	fileWAV := srvTmpFolder + nameFile + ".wav"
	com := fmt.Sprintln(fmt.Sprintf(ffmpegDecodeToWav, file, fileWAV))
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
	AND (CURRENT_TIMESTAMP::date-t1."CallComplete"::date)>t2."Log_Period";`, schemaPG)
	return db.Exec(query).Error
}

func isEnabled(db *gorm.DB, name string) (CallingSysSettings, error) {
	var sys CallingSysSettings
	if err := db.Where(`"Enabled"='true' AND "SystemName"=?`, name).Find(&sys).Error; err != nil {
		return sys, err
	}
	return sys, nil
}

func updateAPI(db *gorm.DB, model interface{}, sys CallingSysSettings) *gorm.DB {
	return db.Model(model).Updates(map[string]interface{}{"url": sys.Address, "user": sys.AuthName, "pass": sys.AuthKey}).Take(model)
}

func (i CallingSysTestResults) updateCallsInfo(db *gorm.DB, callID string) error {
	if err := db.Model(&i).Where(`"CallID"=?`, callID).Updates(i).Error; err != nil {
		return err
	}
	return nil
}

func insertEmptyFiles(db *gorm.DB, callID string) error {
	label := "C&V:test system didn't provide audio files"
	callsinfo := CallingSysTestResults{
		DataLoaded:  true,
		AudioFile:   []byte(label),
		AudioGraph:  labelEmptyBMP(label),
		ConnectTime: 0,
	}
	if err := callsinfo.updateCallsInfo(db, callID); err != nil {
		return err
	}
	return nil
}

func labelEmptyBMP(label string) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 500, 100))
	for x := 0; x < 500; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.White)
		}
	}
	x, y := 85, 50
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{255, 0, 0, 255}),
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)},
	}
	d.DrawString(label)

	var buff bytes.Buffer
	bmp.Encode(&buff, img)
	return buff.Bytes()
}

// CallsStatistics returns a structure for adding statistics to the PurchOppt table
// For each test by its testID, I find the number of calls [Count(&total)],
// find the end time of the last call [MAX("CallComplete")]
// Counting the number of calls with a duration> 0,
// and summarize the total duration of these calls
func callsStatistic(db *gorm.DB, testid string) PurchOppt {
	var total, complete, sumcalls float64
	var max time.Time
	var tr CallingSysTestResults
	db.Model(&tr).
		Where(`"CallListID" = ?`, testid).
		Select(`MAX("CallComplete")`).
		Count(&total).
		Row().
		Scan(&max)
	db.Model(&tr).
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
		TestMinutes:  sumcalls / 60,
	}
	return stat
}

func (po PurchOppt) failedTest(db *gorm.DB, request int, comment string) {
	// po.TestingSystemRequestID = "-1"
	po.TestedUntil = time.Now()
	po.TestComment = comment
	db.Model(&po).Where(`"RequestID"=?`, request).Update(po)
}

func (po PurchOppt) updateTestInfo(db *gorm.DB, id int) error {
	return db.Model(&po).Where(`"RequestID"=?`, id).Update(po).Error
}

func (po PurchOppt) updateStatistic(db *gorm.DB, id string) error {
	return db.Model(&po).Where(`"TestingSystemRequestID"=?`, id).Update(po).Error
}

func iTestParseTime(strTime string) time.Time {
	var t time.Time
	if strTime == "" {
		return t
	}
	st := strings.Split(strTime, ".")
	t, _ = time.Parse("15:04:05", st[0])
	return t
}

func assureParseTime(strTime string) time.Time {
	var t time.Time
	if strTime == "" {
		return t
	}
	st := strings.Split(strTime, ".")
	t, _ = time.Parse("2006-01-02T15:04:05", st[0])
	return t
}

func netsenseParseTime(strTime string) time.Time {
	var t time.Time
	if strTime == "" {
		return t
	}
	st := strings.Split(strTime, ".")
	t, _ = time.Parse("2006-01-02 15:04:05", st[0])
	return t
}

func createGZ(content []byte, nameFile string) error {
	fileGZ := srvTmpFolder + nameFile + ".amr.gz"
	file, err := os.Create(fileGZ)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	if err != nil {
		return err
	}
	return nil
}

func uncompressGZ(name string) error {
	pathGZ := srvTmpFolder + name + ".amr.gz"
	gzipFile, err := os.Open(pathGZ)
	if err != nil {
		return err
	}
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	unzipNameFile := strings.Split(pathGZ, ".gz")
	outfileWriter, err := os.Create(unzipNameFile[0])
	if err != nil {
		return err
	}
	io.Copy(outfileWriter, gzipReader)
	outfileWriter.Close()
	gzipReader.Close()
	gzipFile.Close()

	return nil
}

func ignoreWrongNode(resp io.ReadCloser) (io.Reader, error) {
	scan := bufio.NewScanner(resp)
	var i int
	var newResp string = `<?xml version="1.0" encoding="UTF-8" ?>`
	for scan.Scan() {
		i++
		if i == 1 || i == 3 || i == 12 {
			continue
		}
		str := strings.ReplaceAll(scan.Text(), "&", "&amp;")
		newResp = newResp + fmt.Sprintf("%s", strings.TrimSpace(str))
	}

	if err := scan.Err(); err != nil {
		return nil, err
	}

	return strings.NewReader(newResp), nil
}

func xmlDecoder(res *http.Response) *xml.Decoder {
	decoder := xml.NewDecoder(res.Body)
	decoder.Strict = false
	decoder.CharsetReader = charset.NewReaderLabel
	return decoder
}

func createFile(rc io.ReadCloser, nameFile string) error {
	filepath := srvTmpFolder + nameFile
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, rc)
	if err != nil {
		return err
	}
	// _, err = ioutil.ReadFile(filepath)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func concatMP3files(callID string) error {
	//call-20200203123456789-r.mp3 or call-20200203123456789.mp3
	pathBeep := srvTmpFolder + "call-" + callID + "-r.mp3"
	pathAnsw := srvTmpFolder + "call-" + callID + ".mp3"
	pathOut := srvTmpFolder + "out-" + callID + ".mp3"
	com := fmt.Sprintln(fmt.Sprintf(ffmpegConcatMP3, pathBeep, pathAnsw, pathOut))
	_, err := execCommand(com)
	if err != nil {
		return err
	}
	return nil
}

func calcCoordinate(callID string) (float64, int, error) {
	var files [2]string
	files[0] = srvTmpFolder + "call-" + callID + "-r.mp3"
	files[1] = srvTmpFolder + "call-" + callID + ".mp3"
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
