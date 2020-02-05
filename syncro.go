// syncro.go
//
// The file contains data synchronization functions
// between clients using test systems and Captura.
//
package main

import (
	"encoding/json"
	"io"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
	log "redits.oculeus.com/asorokin/CaptTestCallsSrvc/logger"
)

func checkNewSync(db *gorm.DB, api tester, interval int64) {
	sysID := api.sysID(db)
	sysName := api.sysName(db)
	log.Info("Start function syncronization for test system", sysName)
	var sync syncAutomation
	for {
		err := db.Model(&sync).Where("systemid=? AND do_synch=true", sysID).
			Updates(syncAutomation{
				Syncstate: 1,
				SyncStart: time.Now(),
				Comment:   "Start syncronization"}).First(&sync).Error
		switch {
		case err != nil:
			if err.Error() == "record not found" {
				log.Debug("Не было команды синхронизации для", sysName)
			}
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}
		log.Infof("Run sinchronization %s for test system %s", sync.Synctype, sysName)
		if err := api.runSyncro(db, sync); err != nil {
			log.Error(999, "Ошибка синхронизации", err)
			// ! Изменение статуса и коммента
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

func (api assureAPI) runSyncro(db *gorm.DB, s syncAutomation) error {

	switch s.Synctype {
	case "assure_routes_sync":
		if err := api.getSyncroRoutes(db); err != nil {
			return err
		}
	case "assure_destinations_sync":
		if err := api.getSyncroDestinations(db); err != nil {
			return err
		}
	}
	return nil
}

func (api netSenseAPI) runSyncro(db *gorm.DB, s syncAutomation) error {
	log.Info("Start sinchronization for test system", api.SystemName)
	return nil
}

func (api itestAPI) runSyncro(db *gorm.DB, s syncAutomation) error {
	log.Info("Start sinchronization for test system", api.SystemName)
	return nil
}

func (api assureAPI) getSyncroRoutes(db *gorm.DB) error {
	log.Debug("Downloading data for updating routes for system", api.SystemName)
	// log.Debug("API Settings", api)
	var err error
	request := "assure_routes"
	res, err := api.requestGET(api.QueryResults + api.Routes)
	// log.Debug("Prepare response", res)
	if err != nil {
		// log.Errorf(600, "Failed to get a response to the request %s|%v", request, err)
		return err
	}
	log.Debug("Successful response to the request", request)
	start := time.Now()
	log.Debug("Start transaction insert into the table", request)
	if err := insertAssureData(db, request, res.Body); err != nil {
		// log.Errorf(601, "Could not insert data from response %s|%v", request, err)
		return err
	}
	res.Body.Close()
	log.Debug("Successfully insert data from response", request)
	log.Debugf("Elapsed time transaction insert %s %v", request, time.Since(start))

	// log.Infof("The next data update to prepare %s after %d hours", api.SystemName, interval)
	return nil
}

func (api assureAPI) getSyncroDestinations(db *gorm.DB) error {
	log.Debug("Downloading data for updating destinations for system", api.SystemName)
	// log.Debug("API Settings", api)

	request := "assure_destinations"
	var err error
	res, err := api.requestGET(api.QueryResults + api.Destinations)
	// log.Debug("Prepare response", res)
	if err != nil {
		// log.Errorf(600, "Failed to get a response to the request %s|%v", request, err)
		return err
	}
	log.Debug("Successful response to the request", request)
	start := time.Now()
	log.Debug("Start transaction insert into the table", request)
	if err := insertAssureData(db, request, res.Body); err != nil {
		// log.Errorf(601, "Could not insert data from response %s|%v", request, err)
		return err
	}
	res.Body.Close()
	log.Debug("Successfully insert data from response", request)
	log.Debugf("Elapsed time transaction insert %s %v", request, time.Since(start))

	// log.Infof("The next data update to prepare %s after %d hours", api.SystemName, interval)
	return nil
}

func insertAssureData(db *gorm.DB, req string, body io.ReadCloser) error {
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
	case "assure_routes":
		var route assureRoute // table's struct
		var routes routes     // JSON's struct
		if err := db.Delete(assureRoute{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", route.TableName())

		if err := json.NewDecoder(body).Decode(&routes); err != nil {
			return err
		}
		for _, assureRoute := range routes.QueryResult1 {
			if err := mapstructure.Decode(assureRoute, &route); err != nil {
				return err
			}
			if dialectDB == "sqlite3" {
				if err := tx.Create(&route).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, route)
		}
	case "assure_destinations":
		var dest assureDestination
		var dests destinations
		if err := db.Delete(assureDestination{}).Error; err != nil {
			return err
		}
		log.Debug("Successeful truncate table", dest.TableName())

		if err := json.NewDecoder(body).Decode(&dests); err != nil {
			return err
		}
		for _, assureDest := range dests.QueryResult1 {
			if err := mapstructure.Decode(assureDest, &dests); err != nil {
				return err
			}
			if dialectDB == "sqlite3" {
				if err := tx.Create(&dest).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
			bulkslice = append(bulkslice, dest)
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
