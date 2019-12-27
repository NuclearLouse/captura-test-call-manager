// netsense.go
//
// The file contains the basic functions necessary for the operation of the "Netsense" system.
//
// Must be present functions that satisfy the tester interface
//
package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "redits.oculeus.com/asorokin/my_packages/logging"
)

func (api *netSenseAPI) sysName(db *gorm.DB) string {
	db.Take(api)
	return api.SystemName
}

func (api netSenseAPI) prepareRequests(db *gorm.DB, interval int64) {
	log.Info("Send preparatory requests for", api.SystemName)
	log.Debug("API Settings", api)
	for {
		log.Infof("The next data update to prepare %s after %d hours", api.SystemName, interval)
		time.Sleep(time.Duration(interval) * time.Hour)
	}

}

func (api netSenseAPI) runNewTest(db *gorm.DB, nt foundTest) error {
	return nil
}

func (api netSenseAPI) checkTestComplete(db *gorm.DB, launchTest foundTest) error {
	return nil
}

func (api netSenseAPI) uploadResultFiles(db *gorm.DB) {
	return
}

func (api netSenseAPI) destinationsync(db *gorm.DB) error {
	url := "https://netsense.arptel.com/netsense/" // из API
	req := "destinationsync/"                      // из API
	myurl := url + req
	AuthKey := "8e66ab96dd031d83308212f0567ae86b" // из API
	timeZone := "Europe/Stockholm"                // из settings?
	isLastRequest := "true"                       // из settings?
	header := fmt.Sprintf(`<destinationRequest>
	<authentication>
	  <key>%s</key>
	</authentication>
	<parameters>
	  <destinationList>`, AuthKey)
	var l []destinationList
	var body string
	for i := range l {
		d := fmt.Sprintf(`<list>
			<destination>%s</destination>
			<code>%s</code>
			<alias>%s</alias>
		  </list>`, l[i].destination, l[i].code, l[i].alias)
		body = body + d
	}
	footer := fmt.Sprintf(`</destinationList>
		</parameters>
		<settings>
		  <timeZone>%s</timeZone>
		  <isLastRequest>%s</isLastRequest>
		  <chainCode/>
		</settings>
	  </destinationRequest>`, timeZone, isLastRequest)

	xmlbody := header + body + footer
	resp, err := http.Post(myurl, "application/xml", strings.NewReader(xmlbody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil

	//! если их серевер при запросе будет требовать в заголовке поле Accept,
	//! то запрос буду составлять по другому
	//   // or you can use []byte(`...`) and convert to Buffer later on
	//   body := "<request> <parameters> <email>test@test.com</email> <password>test</password> </parameters> </request>"

	//   client := &http.Client{}
	//   // build a new request, but not doing the POST yet
	//   req, err := http.NewRequest("POST", "http://localhost:8080/", bytes.NewBuffer([]byte(body)))
	//   if err != nil {
	// 	  fmt.Println(err)
	//   }
	//   // you can then set the Header here
	//   // I think the content-type should be "application/xml" like json...
	//   req.Header.Add("Content-Type", "application/xml; charset=utf-8")
	//   // now POST it
	//   resp, err := client.Do(req)
	//   if err != nil {
	// 	  fmt.Println(err)
	//   }
	//   fmt.Println(resp)
}

func (api netSenseAPI) routesync(db *gorm.DB) error {
	url := "https://netsense.arptel.com/netsense/" // из API
	req := "routesync/"
	myurl := url + req
	AuthKey := "8e66ab96dd031d83308212f0567ae86b"
	parentNode := " Parent Folder Path "                 //еще не знаю что это
	dialerGroup := " Dialer Group Path "                 //еще не знаю что это
	filter := "NO/GROUPED-ABC/DETAILED-ABC (Default NO)" //еще не знаю что это
	header := fmt.Sprintf(`<routeRequest>
	<authentication>
	  <key>%s</key>
	</authentication>
	<settings>
	  <parentNode>%s</parentNode> 
	  <dialerGroup>%s</dialerGroup>
	  <filter>%s</filter>
	</settings>
	<parameters>
	  <routeList>`, AuthKey, parentNode, dialerGroup, filter)
	var l []routeList
	var body string
	for i := range l {
		d := fmt.Sprintf(`<list>
        <externalId>%s</externalId>
        <folder>%s</folder>
        <carrier>%s</carrier>
        <trunkGroup>%s</trunkGroup>
        <trunk>%s</trunk>
        <description>%s</description>
        <dialerGroup>%s</dialerGroup>
        <iac>%s</iac>
        <nac>%s</nac>
      </list>`, l[i].externalID, l[i].folder, l[i].carrier, l[i].trunkGroup, l[i].trunk, l[i].description, l[i].dialerGroup, l[i].iac, l[i].nac)
		body = body + d
	}
	footer := "</routeList></parameters></routeRequest>"

	xmlbody := header + body + footer

	resp, err := http.Post(myurl, "application/xml", strings.NewReader(xmlbody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
