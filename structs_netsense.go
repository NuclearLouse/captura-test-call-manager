// structs_netsense_db.go
//
// The file contains the structures that are needed to work with "Netsense" tables.
//
package main

import (
	"encoding/xml"
)

// type destinationList struct {
// 	destination string
// 	code        string
// 	alias       string
// }

// type routeList struct {
// 	externalID  string // External ID
// 	folder      string // Folder (OPTIONAL)
// 	carrier     string // Carrier (OPTIONAL)
// 	trunkGroup  string // Trunk Group (OPTIONAL)
// 	trunk       string // Trunk Name /trunk string
// 	description string // Description (OPTIONAL)
// 	dialerGroup string // Dialer Group (OPTIONAL)
// 	iac         string // IAC (OPTIONAL)
// 	nac         string // NAC (OPTIONAL)
// }

type status struct {
	XMLName  xml.Name `xml:"callListIdRequest"`
	CallList string   `xml:"callListIdList"`
}

type statusResponse struct {
	XMLName               xml.Name `xml:"callListResponseIdArray"`
	CallListResponseArray struct {
		CallListID       string `xml:"callListId"`
		CallLogResponses struct {
			CallListLogID string `xml:"callListLogId"`
			Status        string `xml:"status"`
		} `xml:"callLogResponses"`
	} `xml:"callListResponseArray"`
}

func (netSenseAPI) TableName() string {
	return schemaPG + "CallingSys_API_NetSense"
}

type netSenseAPI struct {
	SystemName       string `gorm:"size:50;foreignkey:CallingSys_Settings.SystemName"`
	SystemID         int    `gorm:"type:int"`
	URL              string `gorm:"size:100"`
	User             string `gorm:"size:100"`
	AuthKey          string `gorm:"size:100"`
	TestInit         string `gorm:"size:100"`
	TestInitList     string `gorm:"size:100"`
	GetCallListid    string `gorm:"size:100"`
	GetDestinations  string `gorm:"size:50"`
	SyncDestinations string `gorm:"size:50"`
	SyncRoutes       string `gorm:"size:50"`
	AudioFile        string `gorm:"size:50"`
	GetAllCalls      string `gorm:"size:50"`
	Version          string `gorm:"size:50"`
}

type testInit struct {
	XMLName    xml.Name `xml:"callRequest"`
	Auth       auth
	Parameters parameters
	Settings   settings
}

type auth struct {
	XMLName xml.Name `xml:"authentication"`
	Key     string   `xml:"key"`
}

type settings struct {
	XMLName      xml.Name `xml:"settings"`
	TimeZone     string   `xml:"timeZone"`
	WebServiceID int      `xml:"webServiceId"`
}

type parameters struct {
	XMLName          xml.Name `xml:"parameters"`
	RoutesList       routesList
	DestinationsList destinationsList
	CallTypeList     callTypeList
}

type routesList struct {
	XMLName xml.Name `xml:"routeList"`
	List    []routeList
}

type routeList struct {
	XMLName xml.Name `xml:"list"`
	Route   string   `xml:"route"`
}

type destinationsList struct {
	XMLName xml.Name `xml:"destinationList"`
	List    []destList
}

type destList struct {
	XMLName     xml.Name `xml:"list"`
	Destination string   `xml:"destination"`
}

type callTypeList struct {
	XMLName xml.Name `xml:"callTypeList"`
	List    []typeList
}

type typeList struct {
	XMLName  xml.Name `xml:"list"`
	CallType string   `xml:"callType"`
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

// <subStatus>
// 	<list>
// 		<status>
// 			<code>5001</code>
// 			<message>Could not save call list</message>
// 		</status>
// 	</list>
// </subStatus>

// type list struct {
// 	XMLName xml.Name `xml:"list"`
// 	Status  status
// }

// type status struct {
// 	XMLName xml.Name `xml:"status"`
// 	Code    string   `xml:"code"`
// 	Message string   `xml:"message"`
// }

// type callResult struct {
// 	XMLName           xml.Name `xml:"callResult"`
// 	AlertTime         float64  `xml:"alertTime"`
// 	CallComplete      string   `xml:"callComplete"` //time.Time
// 	CallCompleteMs    int      `xml:"callCompleteMs"`
// 	CallDuration      float64  `xml:"callDuration"`
// 	CallListLogID     int      `xml:"callListLogId"`
// 	CallListRepeatNr  int      `xml:"callListRepeatNr"`
// 	CallListRowNr     int      `xml:"callListRowNr"`
// 	CallResultID      int      `xml:"callResultId"`
// 	CallStart         string   `xml:"callStart"` //time.Time
// 	CallStartMs       int      `xml:"callStartMs"`
// 	CallType          string   `xml:"callType"`
// 	CallingNumber     string   `xml:"callingNumber"`
// 	CauseCodeID       int      `xml:"causeCodeId"`
// 	CauseLocationID   int      `xml:"causeLocationId"`
// 	ConnectTime       float64  `xml:"connectTime"`
// 	Country           string   `xml:"country"`
// 	Destination       string   `xml:"destination"`
// 	DialedPhoneNumber string   `xml:"dialedPhoneNumber"`
// 	Dialer            string   `xml:"dialer"`
// 	DialerGroup       string   `xml:"dialerGroup"`
// 	DisconnectTime    float64  `xml:"disconnectTime"`
// 	ExternalBatchID   string   `xml:"externalBatchId"`
// 	ExternalCallID    string   `xml:"externalCallId"`
// 	Label             string   `xml:"label"`
// 	LastDigitSent     float64  `xml:"lastDigitSent"`
// 	Pdd               float64  `xml:"pdd"`
// 	PhoneNumber       string   `xml:"phoneNumber"`
// 	PhoneNumberName   string   `xml:"phoneNumberName"`
// 	PhoneNumberType   string   `xml:"phoneNumberType"`
// 	Region            string   `xml:"region"`
// 	Route             string   `xml:"route"`
// 	RouteFolder       string   `xml:"routeFolder"`
// 	Status            string   `xml:"status"`
// 	StatusCode        int      `xml:"statusCode"`
// 	ToneDetection     float64  `xml:"toneDetection"`
// 	Username          string   `xml:"username"`
// 	WebServiceID      int      `xml:"webServiceId"`
// 	AudioURL          string   `xml:"audioUrl"`
// }
