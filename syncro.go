// syncro.go
//
// The file contains data synchronization functions
// between clients using test systems and Captura.
//
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jinzhu/gorm"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
	log "redits.oculeus.com/asorokin/CaptTestCallsSrvc/logger"
)

func checkNewSync(db *gorm.DB, api tester, interval int64) {
	sysID := api.sysID(db)
	sysName := api.sysName(db)
	log.Info("Start service syncronization for test system", sysName)
	for {
		var sync syncAutomation
		err := db.Model(&sync).Where("systemid=? AND do_synch=true", sysID).
			Updates(syncAutomation{
				Syncstate: 1,
				SyncStart: time.Now(),
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
			log.Error(999, "Error syncronization", err)
			message := fmt.Sprintf("Syncronization ERROR:%v", err)
			db.Model(&sync).Where("systemid=? AND do_synch=true", sysID).
				Updates(map[string]interface{}{
					"syncstate": 3,
					"do_synch":  false,
					"sync_end":  time.Now(),
					"comment":   message})
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		db.Model(&sync).Where("systemid=? AND do_synch=true", sysID).
			Updates(map[string]interface{}{
				"syncstate": 3,
				"do_synch":  false,
				"sync_end":  time.Now(),
				"comment":   "Syncronization complete"})
		log.Infof("Sinchronization %s complete for test system %s", sync.Synctype, sysName)
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
		if err := callSyncRoutesFunction(db, api.SystemID); err != nil {
			return err
		}
	case "assure_destinations":
		if err := api.getAssureSynchro(db, api.Destinations); err != nil {
			return err
		}
	}
	return nil
}

func (api assureAPI) getAssureSynchro(db *gorm.DB, r string) error {
	log.Infof("Start downloading data for updating %s for system %s", r, api.SystemName)
	// log.Debug("API Settings", api)
	var err error
	res, err := api.requestGET(api.QueryResults + r)
	if err != nil {
		return err
	}
	log.Debug("Successful response to the request", r)
	//! ***************For DEBUG save response json*********************
	// body, _ := ioutil.ReadAll(res.Body)
	// nameFile := fmt.Sprintf("c:\\capturasystem\\TestCallsManagement_Log\\%s.json", r)
	// ioutil.WriteFile(nameFile, body, 0666)
	//! ****************************************************************
	start := time.Now()
	log.Infof("Start transaction insert into the table CallingSys_assure_%s", r)
	if err := insertAssureData(db, r, res.Body); err != nil {
		return err
	}
	res.Body.Close()
	log.Infof("Successfully insert data. Elapsed time transaction %s %v", r, time.Since(start))
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
