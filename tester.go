// tester.go
//
// The file contains an interface "tester" declaration that all test systems must satisfy.
// For each active system that satisfies the interface, the functions necessary
// for its operation are launched in separate threads.
//
package main

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	log "redits.oculeus.com/asorokin/my_packages/logging"
)

type tester interface {
	sysName(*gorm.DB) string
	sysID(*gorm.DB) int
	runNewTest(*gorm.DB, foundTest) error
	runSyncro(*gorm.DB, syncAutomation) error
	checkTestComplete(*gorm.DB, foundTest) error
	// checkAuth(*gorm.DB) bool
}

func runService(cfg *Config, db *gorm.DB) {
	log.Info("*************Start service*************")

	go checkOldTests(cfg, db)

	ts := []tester{&itestAPI{}, &netSenseAPI{}, &assureAPI{}}
	for i := range ts {
		sysname := ts[i].sysName(db)
		sys, err := isEnabled(db, sysname)
		if err != nil {
			continue
		}
		if err := updateAPI(db, ts[i], sys).Error; err != nil {
			log.Errorf(3, "Cann't update API settings for the system %s|%v", sysname, err)
			continue
		}
		log.Info("Active test system", sysname)
		// switch ts[i].checkAuth(db) {
		// case true:
		// 	if cfg.Application.PrepareRequest {
		// 		go ts[i].prepareRequests(db, cfg.Application.IntervalPrepareTests)
		// 	}
		// 	go checkTestStatus(db, ts[i], cfg.Application.IntervalCheckTests)

		// case false:
		// 	log.Errorf(1, "Authentication failed! Check your internet or database connection and username or password for Test System %s", sysname)
		// }
		// if cfg.Application.PrepareRequest {
		// 	go ts[i].prepareRequests(db, cfg.Application.IntervalPrepareTests)
		// }
		go checkNewSync(db, ts[i], cfg.Application.IntervalCheckSyncro)
		go checkTestStatus(db, ts[i], cfg.Application.IntervalCheckTests)

	}
}

// The main function that works at a given interval and checks for new tests
func checkTestStatus(db *gorm.DB, api tester, interval int64) {
	for {
		sysName := api.sysName(db)
		// po="Purch_Oppt" | ps="Purch_Statuses"| ss="CallingSys_Settings"| rt="CallingSys_RouteList"
		query := fmt.Sprintf(`SELECT po."Request_by_User",
		po."RequestID",
		COALESCE(po."TestingSystemRequestID",'') "TestingSystemRequestID",
		po."RequestState", 
		po."Route_Carrier", 
		rt."Remote_Route_ID",
		po."SupplierID",
		po."Test_Calls",
		COALESCE(po."Test_Comment",'') "Test_Comment",
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
		AND ss."SystemName"='%s' AND po."RequestState" < 3;`, schemaPG, sysName)

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
				&test.TestComment,
				&test.BNumber,       //Custom_BNumbers
				&test.Destination,   //Destination
				&test.DestinationID, //remote_destination_id from CallingSys_DestinationList
				&test.SystemID,      //SystemID from CallingSys_Settings
				&test.SystemName,    //SystemName from CallingSys_Settings
				&test.TestType)      //TestSystemCallType from Purch_Statuses
			if err != nil {
				log.Errorf(10, "Could not add individual tests to the list of tests found for system %s|%v", sysName, err)
				newTestInfo := purchOppt{
					TestingSystemRequestID: "-1",
					TestedUntil:            time.Now(),
					TestComment:            err.Error()}
				if err := newTestInfo.updateTestInfo(db, test.RequestID); err != nil {
					log.Errorf(1, "Cann't insert data about 'not add individual test to the test list'|%v", err)
				}
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
					log.Errorf(7, "Could not start a new test for system %s|%v", t.SystemName, err)
					newTestInfo := purchOppt{
						TestingSystemRequestID: "-1",
						TestedUntil:            time.Now(),
						TestComment:            err.Error()}
					if err := newTestInfo.updateTestInfo(db, t.RequestID); err != nil {
						log.Errorf(1, "Cann't insert data about 'not start new test'|%v", err)
					}
					continue
				}
			case 2:
				log.Infof("Checking the end of test %s for system %s and test_id:%s", t.TestType, t.SystemName, t.TestingSystemRequestID)
				if err := api.checkTestComplete(db, t); err != nil {
					log.Errorf(8, "Could not check status for test %s system %s|%v", t.TestingSystemRequestID, t.SystemName, err)
					newTestInfo := purchOppt{
						TestingSystemRequestID: "0",
						TestedUntil:            time.Now(),
						TestComment:            err.Error()}
					if err := newTestInfo.updateTestInfo(db, t.RequestID); err != nil {
						log.Errorf(1, "Cann't insert data about 'check test status'|%v", err)
					}
					continue
				}
			}
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

//Function that works at a specified interval and removes old tests
func checkOldTests(cfg *Config, db *gorm.DB) {
	for {
		log.Info("Start function delete old test info")
		query := fmt.Sprintf(`DELETE FROM %s"CallingSys_TestResults" AS t1 
		USING %[1]s"CallingSys_Settings" AS t2 
		WHERE t1."TestSystem"=t2."SystemID" 
		AND (CURRENT_TIMESTAMP::date-t1."CallComplete"::date)>t2."Log_Period";`, schemaPG)
		if err := db.Exec(query).Error; err != nil {
			log.Errorf(2, "Error delete old test info|%v", err)
		}
		log.Infof("Next delete old tests info after %d hours", cfg.Application.IntervalDeleteTests)

		// For the sake of variety, I decided to try using a timer rather than the Sleep function
		timer := time.NewTimer(time.Duration(cfg.Application.IntervalDeleteTests) * time.Hour)
		<-timer.C
	}
}
