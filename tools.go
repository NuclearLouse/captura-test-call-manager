// tools.go
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
	for _, table := range listTables {
		// if err := db.AutoMigrate(table).Error; err != nil {
		// 	log.Errorf(9, "Cann't create table|%v", err)
		// 	errs = err
		// }
		if !db.HasTable(table) {
			if err := db.CreateTable(table).Error; err != nil {
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

func deleteFiles(nameFiles []string) {
	var err error
	for _, file := range nameFiles {
		if err = os.Remove(file); err != nil {
			log.Error(4, "Cann't delete file", file)
			continue
		}
		log.Debug("Successefuly delete file", file)
	}
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

// The function draws the waveform of the wav file.
// A png file will be created, a vertical line will be drawn in it at the given x coordinate,
// then the png file will be decoded in bmp. The file will be read in a slice of bytes,
// and all auxiliary files wav, png, bmp will be deleted.
func waveFormImage(nameFile string, x int) ([]byte, error) {
	pathWavFile := srvTmpFolder + nameFile + ".wav"
	pathPngFile := srvTmpFolder + nameFile + ".png"

	if _, err := os.Stat(pathPngFile); !os.IsNotExist(err) {
		if err := os.Remove(pathPngFile); err != nil {
			log.Error(4, "Cann't delete file", pathPngFile)
		}
	}
	com := fmt.Sprintln(fmt.Sprintf(ffmpegWavFormImg, pathWavFile, pathPngFile))
	_, err := execCommand(com)
	if err != nil {
		return nil, err
	}

	// drawing a vertical red line indicating the beginning of the answer
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
	listDeleteFiles := []string{
		pathWavFile,
		pathPngFile,
		pathBmpFile}
	deleteFiles(listDeleteFiles)
	return content, nil
}

// The function draws a vertical line in the given coordinate
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

// Function decodes png to bmp
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

// Function decoding any audio file from the specified codec (mp3 or amr) to wav.
// After that, the wav file is read in a slice of bytes, and the encoded file is deleted.
func decodeToWAV(nameFile, codec string) ([]byte, error) {
	file := srvTmpFolder + nameFile + "." + codec
	fileWAV := srvTmpFolder + nameFile + ".wav"

	if _, err := os.Stat(fileWAV); !os.IsNotExist(err) {
		if err := os.Remove(fileWAV); err != nil {
			log.Error(4, "Cann't delete file", fileWAV)
		}
	}
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

	if err := os.Remove(file); err != nil {
		log.Error(4, "Cann't delete file", file)
	}
	return content, nil
}

// Function checks if the test system is active.
func isEnabled(db *gorm.DB, name string) (CallingSysSettings, error) {
	var sys CallingSysSettings
	if err := db.Where(`"Enabled"='true' AND "SystemName"=?`, name).Find(&sys).Error; err != nil {
		return sys, err
	}
	return sys, nil
}

// Function updating the address, username and password for the test system
func updateAPI(db *gorm.DB, model interface{}, sys CallingSysSettings) *gorm.DB {
	return db.Model(model).Updates(map[string]interface{}{"url": sys.Address, "user": sys.AuthName, "pass": sys.AuthKey}).Take(model)
}

// The function adds a row to the CallingSys_TestResults table for the given call_id
func (i CallingSysTestResults) updateCallsInfo(db *gorm.DB, callID string) error {
	if err := db.Model(&i).Where(`"CallID"=?`, callID).Updates(i).Error; err != nil {
		return err
	}
	return nil
}

// If the test does not provide an audio file, the function will add “empty” information
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

// The function draws a bmp label with the inscription that the test did not provide an audio file
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

// The function updates the information in the Purch_Oppt table about the running test.
func (po PurchOppt) updateTestInfo(db *gorm.DB, id int) error {
	return db.Model(&po).Where(`"RequestID"=?`, id).Update(po).Error
}

// The function updates the statistic information in the Purch_Oppt table for ended test.
func (po PurchOppt) updateStatistic(db *gorm.DB, id string) error {
	return db.Model(&po).Where(`"TestingSystemRequestID"=?`, id).Update(po).Error
}

// The following three functions bring the response fields containing time to a common format
// and type of time, since all systems represent time in different formats.
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
	if err := os.Remove(pathGZ); err != nil {
		log.Error(4, "Cann't delete file", pathGZ)
	}
	return nil
}

// The iTest system, in response to a request for test status,
// returns xml with an incorrect node in the form of a number equal to test_id.
// All Unmarshal functions in this case return with an error.
// Therefore, we had to cut invalid lines. This problem could be solved with the help
// of regular expressions, but this method seemed to me more readable.
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

// The iTest system returns its xml responses not in the standard
// UTF-8 encoding, but in ISO-8859-1, so I had to use an CharsetReader function.
func xmlNewDecoder(res *http.Response) *xml.Decoder {
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
	return nil
}

// The function concatenate two mp3 files ring and answer into one common file.
func concatMP3files(callID string) error {
	//call-20200203123456789-r.mp3 or call-20200203123456789.mp3
	pathRing := srvTmpFolder + "call-" + callID + "-r.mp3"
	pathAnsw := srvTmpFolder + "call-" + callID + ".mp3"
	pathOut := srvTmpFolder + "out-" + callID + ".mp3"

	if _, err := os.Stat(pathOut); !os.IsNotExist(err) {
		if err := os.Remove(pathOut); err != nil {
			log.Error(4, "Cann't delete file", pathOut)
		}
	}
	com := fmt.Sprintln(fmt.Sprintf(ffmpegConcatMP3, pathRing, pathAnsw, pathOut))
	_, err := execCommand(com)
	if err != nil {
		return err
	}
	listDeleteFiles := []string{pathRing, pathAnsw}
	deleteFiles(listDeleteFiles)
	return nil
}

// The function calculates the vertical bar coordinate,
// depending on the duration of the ring and answer audio files.
// And it returns this coordinate as a duration in seconds
// from the start of the call and as the number of pixels.
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
