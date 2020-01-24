// structs_netsense_db.go
//
// The file contains the structures that are needed to work with "Netsense" tables.
//
package main

import (
	"encoding/xml"
)

type destinationList struct {
	destination string
	code        string
	alias       string
}

type routeList struct {
	externalID  string // External ID
	folder      string // Folder (OPTIONAL)
	carrier     string // Carrier (OPTIONAL)
	trunkGroup  string // Trunk Group (OPTIONAL)
	trunk       string // Trunk Name /trunk string
	description string // Description (OPTIONAL)
	dialerGroup string // Dialer Group (OPTIONAL)
	iac         string // IAC (OPTIONAL)
	nac         string // NAC (OPTIONAL)
}

func (netSenseAPI) TableName() string {
	return schemaPG + "CallingSys_API_NetSense"
}

type netSenseAPI struct {
	SystemName       string `gorm:"size:50;foreignkey:CallingSys_Settings.SystemName"`
	URL              string `gorm:"size:100"`
	User             string `gorm:"size:100"`
	AuthKey          string `gorm:"size:100"`
	TestInit         string `gorm:"size:100"`
	GetDestinations  string `gorm:"size:50"`
	SyncDestinations string `gorm:"size:50"`
	SyncRoutes       string `gorm:"size:50"`
	AudioFile        string `gorm:"size:50"`
	GetAllCalls      string `gorm:"size:50"`
	Version          string `gorm:"size:50"`
}

type testInit struct {
	XMLName xml.Name `xml:"callRequest"`
	Auth    struct {
		Key string `xml:"key"`
	} `xml:"authentication"`
	Parameters struct {
		RouteList struct {
			route string
		} `xml:"routeList"`
		DestinationList struct {
			destination string
		} `xml:"destinationList"`
		CallTypeList struct {
			callType string
		} `xml:"callTypeList"`
	} `xml:"parameters"`
	Settings struct {
		TimeZone     string `xml:"timeZone"`
		WebServiceID int    `xml:"webServiceId"`
	} `xml:"settings"`
}

type responseTestInit struct {
	XMLName xml.Name `xml:"response"`
	Status  struct {
		Code    string `xml:"code"`
		Message string `xml:"message"`
	} `xml:"status"`
	SubStatus struct {
		list
	} `xml:"subStatus"`
	callResult
}

type list struct {
	XMLName xml.Name `xml:"list"`
}

type callResult struct {
	XMLName           xml.Name `xml:"callResult"`
	AlertTime         float64  `xml:"alertTime"`
	CallComplete      string   `xml:"callComplete"` //time.Time
	CallCompleteMs    int      `xml:"callCompleteMs"`
	CallDuration      float64  `xml:"callDuration"`
	CallListLogID     int      `xml:"callListLogId"`
	CallListRepeatNr  int      `xml:"callListRepeatNr"`
	CallListRowNr     int      `xml:"callListRowNr"`
	CallResultID      int      `xml:"callResultId"`
	CallStart         string   `xml:"callStart"` //time.Time
	CallStartMs       int      `xml:"callStartMs"`
	CallType          string   `xml:"callType"`
	CallingNumber     string   `xml:"callingNumber"`
	CauseCodeID       int      `xml:"causeCodeId"`
	CauseLocationID   int      `xml:"causeLocationId"`
	ConnectTime       float64  `xml:"connectTime"`
	Country           string   `xml:"country"`
	Destination       string   `xml:"destination"`
	DialedPhoneNumber string   `xml:"dialedPhoneNumber"`
	Dialer            string   `xml:"dialer"`
	DialerGroup       string   `xml:"dialerGroup"`
	DisconnectTime    float64  `xml:"disconnectTime"`
	ExternalBatchID   string   `xml:"externalBatchId"`
	ExternalCallID    string   `xml:"externalCallId"`
	Label             string   `xml:"label"`
	LastDigitSent     float64  `xml:"lastDigitSent"`
	Pdd               float64  `xml:"pdd"`
	PhoneNumber       string   `xml:"phoneNumber"`
	PhoneNumberName   string   `xml:"phoneNumberName"`
	PhoneNumberType   string   `xml:"phoneNumberType"`
	Region            string   `xml:"region"`
	Route             string   `xml:"route"`
	RouteFolder       string   `xml:"routeFolder"`
	Status            string   `xml:"status"`
	StatusCode        int      `xml:"statusCode"`
	ToneDetection     float64  `xml:"toneDetection"`
	Username          string   `xml:"username"`
	WebServiceID      int      `xml:"webServiceId"`
	AudioURL          string   `xml:"audioUrl"`
}
