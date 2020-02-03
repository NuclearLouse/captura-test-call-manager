// structs_netsense_db.go
//
// The file contains the structures that are needed to work with "Netsense" tables.
//
package main

import (
	"encoding/xml"
)

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

//-----------------------------------------------------------------------------
//*******************Block of test initialization structures*******************
//-----------------------------------------------------------------------------
type testInitNetsense struct {
	XMLName        xml.Name `xml:"callListRequest"`
	Authentication authentication
	ParametersList parametersList
}

type authentication struct {
	XMLName xml.Name `xml:"authentication"`
	Key     string   `xml:"key"`
}

type parametersList struct {
	XMLName         xml.Name `xml:"parametersList"`
	CallTypeList    callTypeList
	RouteList       routeList
	DestinationList destinationList
}

type callTypeList struct {
	XMLName xml.Name `xml:"callTypeList"`
	List    listCallType
}

type listCallType struct {
	XMLName  xml.Name `xml:"list"`
	CallType string   `xml:"callType"`
}

type routeList struct {
	XMLName   xml.Name `xml:"routeList"`
	ListRoute listRoute
}

type listRoute struct {
	XMLName xml.Name `xml:"list"`
	Route   string   `xml:"route"`
}

type destinationList struct {
	XMLName xml.Name `xml:"destinationList"`
	List    []listDestination
}

type listDestination struct {
	XMLName     xml.Name `xml:"list"`
	Destination string   `xml:"destination"`
}

type testInitResponse struct {
	XMLName               xml.Name `xml:"callListResponseArray"`
	CallListResponseArray struct {
		Status struct {
			Code    string `xml:"code"`
			Message string `xml:"message"`
		} `xml:"status"`
		SubStatus struct {
			List struct {
				Parameters struct {
					CallTypeList struct {
						List struct {
							CallType string `xml:"callType"`
						} `xml:"list"`
					} `xml:"callTypeList"`
					DestinationList struct {
						List struct {
							Destination string `xml:"destination"`
						} `xml:"list"`
					} `xml:"destinationList"`
				} `xml:"parameters"`
				Status struct {
					Code    string `xml:"code"`
					Message string `xml:"message"`
				} `xml:"status"`
			} `xml:"list"`
		} `xml:"subStatus"`
		ResponseID string `xml:"responseId"`
	} `xml:"callListResponseArray"`
}

//-----------------------------------------------------------------------------
//**************Block of test status and test results structures***************
//-----------------------------------------------------------------------------
type testStatusNetsense struct {
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

type testResultNetsense struct {
	XMLName xml.Name `xml:"response"`
	Result  struct {
		List []struct {
			CallResult struct {
				AlertTime struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"alertTime"`
				AudioURL struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"audioUrl"`
				CallComplete struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callComplete"`
				CallCompleteMs struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callCompleteMs"`
				CallDuration struct {
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"callDuration"`
				CallListLogID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callListLogId"`
				CallListRepeatNr struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callListRepeatNr"`
				CallListRowNr struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callListRowNr"`
				CallResultID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callResultId"`
				CallStart struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callStart"`
				CallStartMs struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callStartMs"`
				CallType struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callType"`
				CallingNumber struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callingNumber"`
				CauseCodeID struct {
					Format string `xml:"format"`
					Value  int    `xml:"value"`
				} `xml:"causeCodeId"`
				CauseLocationID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"causeLocationId"`
				ConnectTime struct {
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"connectTime"`
				Country struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"country"`
				Destination struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"destination"`
				DialedPhoneNumber struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dialedPhoneNumber"`
				Dialer struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dialer"`
				DialerGroup struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dialerGroup"`
				DisconnectTime struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"disconnectTime"`
				DtmfDetected struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"dtmfDetected"`
				ExternalBatchID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"externalBatchId"`
				ExternalCallID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"externalCallId"`
				Label struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"label"`
				LastDigitSent struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"lastDigitSent"`
				MatchPercentage struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"matchPercentage"`
				MediaStartTime struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"mediaStartTime"`
				OntheFlyPhoneNumberName struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"ontheFlyPhoneNumberName"`
				Pdd struct {
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"pdd"`
				PddAlternate struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"pddAlternate"`
				PhoneNumber struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"phoneNumber"`
				PhoneNumberName struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"phoneNumberName"`
				PhoneNumberType struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"phoneNumberType"`
				Region struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"region"`
				RingingDuration struct {
					Format string  `xml:"format"`
					Value  float64 `xml:"value"`
				} `xml:"ringingDuration"`
				Route struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"route"`
				RouteFolder struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"routeFolder"`
				SipCallID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"sipCallId"`
				Status struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"status"`
				StatusCode struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"statusCode"`
				ToneDetection struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"toneDetection"`
				Username struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"username"`
				WebServiceID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"webServiceId"`
			} `xml:"callResult"`
			CallResultCLI struct {
				CLIDetectedCallingNumber struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIDetectedCallingNumber"`
				CLIResult struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIResult"`
				CLIStatus struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIStatus"`
				CLIStatusCode struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIStatusCode"`
				CLIType struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"CLIType"`
				CallResultID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callResultId"`
			} `xml:"callResultCLI"`
			CallResultFAS struct {
				CallResultID struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"callResultId"`
				FasResult struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"fasResult"`
				FasStatus struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"fasStatus"`
				FasStatusCode struct {
					Format string `xml:"format"`
					Value  string `xml:"value"`
				} `xml:"fasStatusCode"`
			} `xml:"callResultFAS"`
		} `xml:"list"`
	} `xml:"result"`
	Status struct {
		Code    string `xml:"code"`
		Message string `xml:"message"`
	} `xml:"status"`
}
