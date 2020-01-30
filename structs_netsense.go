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

type list struct {
	XMLName xml.Name `xml:"list"`
	Status  status
}

type status struct {
	XMLName xml.Name `xml:"status"`
	Code    string   `xml:"code"`
	Message string   `xml:"message"`
}

type testStatus struct {
	XMLName               xml.Name `xml:"callListResponseIdArray"`
	CallListResponseArray struct {
		CallListID       string `xml:"callListId"`
		CallLogResponses struct {
			CallListLogID string `xml:"callListLogId"`
			Status        string `xml:"status"`
		} `xml:"callLogResponses"`
	} `xml:"callListResponseArray"`
}

type getStatus struct {
	XMLName  xml.Name `xml:"callListIdRequest"`
	CallList string   `xml:"callListIdList"`
}

type testResult struct {
	XMLName xml.Name `xml:"response"`
	Text    string   `xml:",chardata"`
	Result  struct {
		Text string `xml:",chardata"`
		List []struct {
			Text       string `xml:",chardata"`
			CallResult struct {
				Text      string `xml:",chardata"`
				AlertTime struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"alertTime"`
				AudioURL struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"audioUrl"`
				CallComplete struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callComplete"`
				CallCompleteMs struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callCompleteMs"`
				CallDuration struct {
					Text   string  `xml:",chardata"`
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"callDuration"`
				CallListLogID struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callListLogId"`
				CallListRepeatNr struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callListRepeatNr"`
				CallListRowNr struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callListRowNr"`
				CallResultID struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callResultId"`
				CallStart struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callStart"`
				CallStartMs struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callStartMs"`
				CallType struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callType"`
				CallingNumber struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callingNumber"`
				CauseCodeID struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"causeCodeId"`
				CauseLocationID struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"causeLocationId"`
				ConnectTime struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"connectTime"`
				Country struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"country"`
				Destination struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"destination"`
				DialedPhoneNumber struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dialedPhoneNumber"`
				Dialer struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dialer"`
				DialerGroup struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dialerGroup"`
				DisconnectTime struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"disconnectTime"`
				DtmfDetected struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dtmfDetected"`
				ExternalBatchID struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"externalBatchId"`
				ExternalCallID struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"externalCallId"`
				Label struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"label"`
				LastDigitSent struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"lastDigitSent"`
				MatchPercentage struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"matchPercentage"`
				MediaStartTime struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"mediaStartTime"`
				OntheFlyPhoneNumberName struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"ontheFlyPhoneNumberName"`
				Pdd struct {
					Text   string  `xml:",chardata"`
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"pdd"`
				PddAlternate struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"pddAlternate"`
				PhoneNumber struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"phoneNumber"`
				PhoneNumberName struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"phoneNumberName"`
				PhoneNumberType struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"phoneNumberType"`
				Region struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"region"`
				RingingDuration struct {
					Text   string  `xml:",chardata"`
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"ringingDuration"`
				Route struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"route"`
				RouteFolder struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"routeFolder"`
				SipCallID struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"sipCallId"`
				Status struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"status"`
				StatusCode struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"statusCode"`
				ToneDetection struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"toneDetection"`
				Username struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"username"`
				WebServiceID struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"webServiceId"`
			} `xml:"callResult"`
			CallResultCLI struct {
				Text                     string `xml:",chardata"`
				CLIDetectedCallingNumber struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIDetectedCallingNumber"`
				CLIResult struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIResult"`
				CLIStatus struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIStatus"`
				CLIStatusCode struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIStatusCode"`
				CLIType struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIType"`
				CallResultID struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callResultId"`
			} `xml:"callResultCLI"`
			CallResultFAS struct {
				Text         string `xml:",chardata"`
				CallResultID struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callResultId"`
				FasResult struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"fasResult"`
				FasStatus struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"fasStatus"`
				FasStatusCode struct {
					Text   string `xml:",chardata"`
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"fasStatusCode"`
			} `xml:"callResultFAS"`
		} `xml:"list"`
	} `xml:"result"`
	Status struct {
		Text    string `xml:",chardata"`
		Code    string `xml:"code"`
		Message string `xml:"message"`
	} `xml:"status"`
}

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
