// service.go
//
// The file contains an interface "tester" declaration that all test systems must satisfy.
// For each active system that satisfies the interface, the functions necessary
// for its operation are launched in separate threads.
//
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"redits.oculeus.com/asorokin/tcm/config"

	log "redits.oculeus.com/asorokin/my_packages/logging"
)

type tester interface {
	sysName(*gorm.DB) string
	prepareRequests(*gorm.DB, int64)
	runNewTest(*gorm.DB, foundTest) error
	checkTestComplete(*gorm.DB, foundTest) error
	uploadResultFiles(*gorm.DB)
}

func runService(cfg *config.Config, db *gorm.DB) {
	log.Info("*************Start service*************")

	ts := []tester{&netSenseAPI{}, &itestAPI{}, &assureAPI{}}
	for i := range ts {
		sysname := ts[i].sysName(db)
		sys, err := isEnabled(db, true, sysname)
		if err != nil {
			continue
		}
		if err := updateAPI(db, ts[i], sys).Error; err != nil {
			log.Errorf(3, "Cann't update API settings for the system %s|%v", sysname, err)
			continue
		}
		log.Info("Active test system", sysname)

		if cfg.Application.PrepareRequest {
			go ts[i].prepareRequests(db, cfg.Timetable.IntervalPrepareTests)
		}

		go checkTestStatus(db, ts[i], cfg.Timetable.IntervalCheckTests)

		go ts[i].uploadResultFiles(db)

	}
	// waitForSignal()
}

func checkTestStatus(db *gorm.DB, api tester, interval int64) {
	for {
		sysName := api.sysName(db)

		query := fmt.Sprintf(`SELECT po."Request_by_User",
		po."RequestID",
		COALESCE(po."TestingSystemRequestID",'') "TestingSystemRequestID",
		po."RequestState", 
		po."Route_Carrier", 
		rt."Remote_Route_ID",
		po."SupplierID",
		po."Test_Calls",
		COALESCE(po."Custom_BNumbers",'') "Custom_BNumbers",
		po."Destination",
		COALESCE(dl.remote_destination_id, -1) remote_destination_id,
		ss."SystemID",
		ss."SystemName",
		ps."TestSystemCallType"
		FROM %s"Purch_Oppt" po
		JOIN %[1]s"Purch_Statuses" ps ON po."Test_Type"=ps."StatusID"
		JOIN %[1]s"CallingSys_DestinationList" dl ON po."DestinationID"=dl.captura_destination_id AND dl.callingsys_id=ps."TestSystem"
		JOIN %[1]s"CallingSys_Settings" ss ON ss."SystemID"=ps."TestSystem"
		LEFT JOIN %[1]s"CallingSys_RouteList" rt ON po."CallingSys_RouteID" = rt."RouteID" 
		WHERE po."Tested_Until" IS NULL 
		AND (po."TestingSystemRequestID" IS NULL OR po."TestingSystemRequestID"<>'-1') 
		AND ss."SystemName"='%s' AND po."RequestState" < 3;`, os.Getenv("SCHEMA_PG"), sysName)

		rows, err := db.Raw(query).Rows()
		if err != nil {
			log.Errorf(6, "Could not select tests to check status for system %s|%v", sysName, err)
			continue
		}
		var ft []foundTest
		for rows.Next() {
			var test foundTest
			err := rows.Scan(&test.RequestByUser, //Request_by_User from Purch_Oppt
				&test.RequestID,              //RequestID from Purch_Oppt
				&test.TestingSystemRequestID, //TestingSystemRequestID
				&test.RequestState,           //RequestState
				&test.RouteCarrier,           //Route_Carrier
				&test.TestSysRouteID,         //CallingSys_RouteID
				&test.SupplierID,             //SupplierID
				&test.TestCalls,              //TestCalls
				&test.BNumber,                //Custom_BNumbers
				&test.Destination,            //Destination
				&test.DestinationID,          //remote_destination_id from CallingSys_DestinationList
				&test.SystemID,               //SystemID from CallingSys_Settings
				&test.SystemName,             //SystemName from CallingSys_Settings
				&test.TestType)               //TestSystemCallType from Purch_Statuses
			if err != nil {
				log.Errorf(10, "Could not add individual tests to the list of tests found for system %s|%v", sysName, err)
				continue
			}
			ft = append(ft, test)
		}

		if len(ft) == 0 {
			log.Debug("No tests found for", sysName)
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}
		for _, t := range ft {
			switch t.RequestState {
			case 1:
				log.Infof("Initiated new test %s for system %s", t.TestType, t.SystemName)
				if err := api.runNewTest(db, t); err != nil {
					log.Errorf(7, "Could not start a new test %s for system %s|%v", t.TestingSystemRequestID, t.SystemName, err)
					newTestInfo := PurchOppt{TestingSystemRequestID: "0"}
					newTestInfo.failedTest(db, t.RequestID, err.Error())
					continue
				}
			case 2:
				log.Infof("Checking the end of test %s for system %s and test_id:%s", t.TestType, t.SystemName, t.TestingSystemRequestID)
				if err := api.checkTestComplete(db, t); err != nil {
					log.Errorf(8, "Could not check status for test %s system %s|%v", t.TestingSystemRequestID, t.SystemName, err)
					newTestInfo := PurchOppt{}
					newTestInfo.failedTest(db, t.RequestID, err.Error())
					continue
				}
			}
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func checkOldTests(cfg *config.Config, db *gorm.DB) {
	for {
		log.Debug("Start function delete old test info")
		if err := deleteOldTestInfo(db); err != nil {
			log.Errorf(2, "Error delete old test info|%v", err)
		}
		log.Debugf("Next delete old tests info after %d hours", cfg.Timetable.IntervalDeleteTests)

		// For the sake of variety, I decided to try using a timer rather than the Sleep function
		timer := time.NewTimer(time.Duration(cfg.Timetable.IntervalDeleteTests) * time.Hour)
		<-timer.C
	}
}
