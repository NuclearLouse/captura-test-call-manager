package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	log "captura_tcm/logger"

	"github.com/jinzhu/gorm"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
)

func checkNewSync(db *gorm.DB, api tester, interval int64) {
	sysID := api.sysID(db)
	sysName := api.sysName(db)
	log.Info("Start service syncronization for test system", sysName)
	for {
		var sync syncAutomation
		start := time.Now()
		err := db.Model(&sync).Where("systemid=? AND do_synch=true", sysID).
			Updates(syncAutomation{
				Syncstate: 2,
				SyncStart: start,
				Comment:   "Start syncronization"}).First(&sync).Error
		switch {
		case err != nil:
			if err.Error() == "record not found" {
				log.Debug("Not command sinchronization for ", sysName)
			}
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}
		log.Infof("Run sinchronization %s for test system %s", sync.Synctype, sysName)
		if err := api.runSyncro(db, sync); err != nil {
			log.Error(5, "Error syncronization", err)
			message := fmt.Sprintf("Syncronization ERROR:%v", err)
			db.Model(&sync).Where("systemid=? AND do_synch=true", sysID).
				Updates(map[string]interface{}{
					"syncstate": -1,
					"do_synch":  false,
					"sync_end":  time.Now(),
					"comment":   message})
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}
		end := time.Since(start)
		message := fmt.Sprintf("Syncronization %s complete. Elapsed time %v", sync.Synctype, end)
		db.Model(&sync).Where("systemid=? AND do_synch=true", sysID).
			Updates(map[string]interface{}{
				"syncstate": 3,
				"do_synch":  false,
				"sync_end":  time.Now(),
				"comment":   message})
		log.Infof("Sinchronization %s complete for test system %s. Elapsed time %v", sync.Synctype, sysName, end)
		time.Sleep(time.Duration(interval) * time.Second)
	}

}

//-----------------------------------------------------------------------------
//*******************Block of Assure syncro functions*******************
//-----------------------------------------------------------------------------
func (api assureAPI) runSyncro(db *gorm.DB, s syncAutomation) error {

	switch s.Synctype {
	case "assure_routes":
		if err := api.getAssureSynchro(db, api.Routes); err != nil {
			return err
		}
		//TODO: ?????? ?????????? ???????????? ??????????????
		if err := callSyncRoutesFunction(db, api.SystemID); err != nil {
			return err
		}
	case "assure_destinations":
		if err := api.getAssureSynchro(db, api.Destinations); err != nil {
			return err
		}
		//TODO: ?????? ?????????? ???????????? ??????????????
		if err := callSyncDestsFunction(db, api.SystemID); err != nil {
			return err
		}
	case "assure_sms_routes":
		if err := api.getAssureSynchro(db, api.SmsRoutes); err != nil {
			return err
		}
		//TODO: ?????? ?????????? ???????????? ??????????????
		if err := callSyncSmsRoutesFunction(db, api.SystemID); err != nil {
			return err
		}
	case "assure_sms_templates":
		if err := api.getAssureSynchro(db, api.SmsTemplates); err != nil {
			return err
		}
		//TODO: ?????? ?????????? ???????????? ??????????????
		if err := callSyncSmsTemplatesFunction(db, api.SystemID); err != nil {
			return err
		}
	}
	return nil
}

func (api assureAPI) getAssureSynchro(db *gorm.DB, r string) error {
	log.Infof("Start downloading data for updating %s for system %s", r, api.SystemName)
	// log.Debug("API Settings", api)
	var err error
	res, err := api.newRequest("GET", api.QueryResults+r, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	log.Debug("Successful response to the request", r)
	start := time.Now()
	log.Debugf("Start transaction insert into the table CallingSys_assure_%s", r)
	if err := insertAssureData(db, r, res.Body); err != nil {
		return err
	}
	log.Debugf("Successfully insert data. Elapsed time transaction %s %v", r, time.Since(start))
	return nil
}

func insertAssureData(db *gorm.DB, r string, body io.ReadCloser) error {
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
	switch r {
	case "Routes":
		var rs routes
		if err := db.Delete(assureRoute{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table CallingSys_assure_routes")

		if err := json.NewDecoder(body).Decode(&rs); err != nil {
			return err
		}
		for _, ar := range rs.QueryResult1 {
			if dialectDB == "sqlite3" {
				if err := tx.Create(&ar).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, ar)
		}
	case "Destinations":
		var ds destinations
		if err := db.Delete(assureDestination{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table CallingSys_assure_destinations")

		if err := json.NewDecoder(body).Decode(&ds); err != nil {
			return err
		}
		for _, ad := range ds.QueryResult1 {
			if dialectDB == "sqlite3" {
				if err := tx.Create(&ad).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, ad)
		}
	case "SMSRoutes":
		var smsr smsRoutes
		if err := db.Delete(assureSmsRoute{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table CallingSys_assure_sms_routes")

		if err := json.NewDecoder(body).Decode(&smsr); err != nil {
			return err
		}
		for _, asr := range smsr.QueryResult1 {
			if dialectDB == "sqlite3" {
				if err := tx.Create(&asr).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, asr)
		}
	case "SMSTemplates":
		var smst smsTemplates
		if err := db.Delete(assureSmsTemplate{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table CallingSys_assure_sms_templates")

		if err := json.NewDecoder(body).Decode(&smst); err != nil {
			return err
		}
		for _, ast := range smst.QueryResult1 {
			if dialectDB == "sqlite3" {
				if err := tx.Create(&ast).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, ast)
		}
	}
	switch dialectDB {
	case "sqlite3":
		err := tx.Commit().Error
		if err != nil {
			return err
		}
	default:
		if err := gormbulk.BulkInsert(db, bulkslice, 1500); err != nil {
			return err
		}
	}

	return nil
}

//-----------------------------------------------------------------------------
//*******************Block of NetSense syncro functions*******************
//-----------------------------------------------------------------------------
func (api netSenseAPI) runSyncro(db *gorm.DB, s syncAutomation) error {
	log.Info("Start sinchronization for test system", api.SystemName)
	return nil
}

//-----------------------------------------------------------------------------
//*******************Block of iTest syncro functions*******************
//-----------------------------------------------------------------------------
func (api itestAPI) runSyncro(db *gorm.DB, s syncAutomation) error {
	log.Info("Start sinchronization for test system", api.SystemName)
	return nil
}
