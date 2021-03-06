package main

import (
	"fmt"
	"syscall"
	"time"

	"github.com/jinzhu/gorm"

	log "captura_tcm/logger"
)

type tester interface {
	checkAuth(*gorm.DB) bool
	sysName(*gorm.DB) string
	sysID(*gorm.DB) int
	runNewTest(*gorm.DB, foundTest) error
	runSyncro(*gorm.DB, syncAutomation) error
	checkTestComplete(*gorm.DB, foundTest) error
	cancelTest(*gorm.DB, string) error
	prepareRequests(*gorm.DB, int64)
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

		if cfg.Application.PrepareRequest {
			go ts[i].prepareRequests(db, cfg.Application.IntervalPrepare)
		}
		go checkNewSync(db, ts[i], cfg.Application.IntervalCheckSyncro)
		go checkTestStatus(db, ts[i], cfg.Application.IntervalCheckTests)

	}
}

func checkTestStatus(db *gorm.DB, api tester, interval int64) {
	for {
		if !api.checkAuth(db) {
			log.Panic(666, "Unauthorized access. Service will be stopped. Verify your username or password and start the service.")
			sigChan <- syscall.SIGTERM
		}
		sysName := api.sysName(db)
		
		query := fmt.Sprintf(`SELECT po."Request_by_User",
		po."RequestID",
		COALESCE(po."TestingSystemRequestID",'') "TestingSystemRequestID",
		po."RequestState", 
		rt."Remote_Route_ID",
		po."SupplierID",
		po."Test_Calls",
		COALESCE(po."Test_Comment",'') "Test_Comment",
		COALESCE(po."Custom_BNumbers",'') "Custom_BNumbers",
		po."Destination",
		COALESCE(dl.remote_destination_id, -1) remote_destination_id,
		COALESCE(ast.name,'') sms_template_name,
		ss."SystemID",
		ss."SystemName",
		ps."TestSystemCallType"
		FROM %s"Purch_Oppt" po
		JOIN %[1]s"Purch_Statuses" ps ON po."Test_Type"=ps."StatusID"
		JOIN %[1]s"CallingSys_DestinationList" dl ON po."DestinationID"=dl.captura_destination_id AND dl.callingsys_id=ps."TestSystem"
		JOIN %[1]s"CallingSys_Settings" ss ON ss."SystemID"=ps."TestSystem"
		JOIN %[1]s"CallingSys_assure_sms_templates" ast ON ast.sms_template_id = po.sms_template_id
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
			err := rows.Scan(
				&test.RequestByUser, //Request_by_User from Purch_Oppt
				&test.RequestID,              //RequestID from Purch_Oppt
				&test.TestingSystemRequestID, //TestingSystemRequestID
				&test.RequestState,           //RequestState
				&test.TestSysRouteID,         //CallingSys_RouteID
				&test.SupplierID,             //SupplierID
				&test.TestCalls,              //TestCalls
				&test.TestComment,
				&test.BNumber,       //Custom_BNumbers
				&test.Destination,   //Destination
				&test.DestinationID, //remote_destination_id from CallingSys_DestinationList
				&test.SMSTemplate,
				&test.SystemID,      //SystemID from CallingSys_Settings
				&test.SystemName,    //SystemName from CallingSys_Settings
				&test.TestType,
				)      //TestSystemCallType from Purch_Statuses
			if err != nil {
				log.Errorf(10, "Could not add individual tests to the list of tests found for system %s|%v", sysName, err)
				if err := testFail(err).updateTestInfo(db, test.RequestID); err != nil {
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
			var newTestInfo purchOppt
			switch t.RequestState {
			case 1:
				log.Infof("Initiated new test %s for system %s", t.TestType, t.SystemName)
				if err := api.runNewTest(db, t); err != nil {
					log.Errorf(7, "Could not start a new test for system %s|%v", t.SystemName, err)
					newTestInfo = testFail(err)
				}

			case 2:
				log.Infof("Checking the end of test %s for system %s and test_id:%s", t.TestType, t.SystemName, t.TestingSystemRequestID)
				if err := api.checkTestComplete(db, t); err != nil {
					log.Errorf(8, "Could not check status for test %s system %s|%v", t.TestingSystemRequestID, t.SystemName, err)
					newTestInfo = testFail(err)
				}

			case -1:
				//! ???????????????????? ???????? ???????????? ?????? ???????????? ????????????
				log.Info("User canceled test")
				if err := api.cancelTest(db, t.TestingSystemRequestID); err != nil {
					log.Errorf(999, "Blb blb blb %s %s|%v", t.TestingSystemRequestID, t.SystemName, err)
				}
				newTestInfo = testCancel()
			}

			if err := newTestInfo.updateTestInfo(db, t.RequestID); err != nil {
				log.Errorf(1, "Cann't insert data in Purch_Oppt table for new test %d|%v", t.RequestID, err)
			}
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

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

		<- time.NewTimer(time.Duration(cfg.Application.IntervalDeleteTests) * time.Hour).C
	}
}
