// structs_callsystem.go
//
// The file contains the structures that are needed to work with Captura database tables.
//
package main

import (
	"strings"
	"time"
)

func (CallingSysSettings) TableName() string {
	return schemaPG + "CallingSys_Settings"
}

type CallingSysSettings struct {
	SystemID    int    `gorm:"column:SystemID;AUTO_INCREMENT"`
	SystemName  string `gorm:"column:SystemName;size:50"`
	Enabled     bool   `gorm:"column:Enabled;default:false"`
	Address     string `gorm:"column:Address;size:100"`
	AuthKey     string `gorm:"column:Auth_Key;size:250"`
	AuthName    string `gorm:"column:Auth_Name;size:100"`
	TestProfile string `gorm:"column:Test_Profile;size:100"`
	LogPeriod   int    `gorm:"column:Log_Period;default:7"`
	TestCodec   string `gorm:"column:Test_Codec;size:5;default:'alaw'"`
	SSL         string `gorm:"column:SSL;size:10;default:'1_2'"`
}

func (CallingSysTestResults) TableName() string {
	return schemaPG + "CallingSys_TestResults"
}

type CallingSysTestResults struct {
	CallID                     string    `gorm:"column:CallID;size:30"`
	CallListID                 string    `gorm:"column:CallListID;size:30"`
	TestSystem                 int       `gorm:"column:TestSystem;type:int;default:0"`
	CallType                   string    `gorm:"column:CallType;size:20"`
	Destination                string    `gorm:"column:Destination;size:100"`
	CallStart                  time.Time `gorm:"column:CallStart;type:timestamp"`
	CallComplete               time.Time `gorm:"column:CallComplete;type:timestamp"`
	CallDuration               float64   `gorm:"column:CallDuration;type:float8;default:0"`
	AlertTime                  float64   `gorm:"column:AlertTime;type:float8"`
	RingDuration               float64   `gorm:"column:RingDuration;type:float8"`
	ConnectTime                float64   `gorm:"column:ConnectTime;type:float8"`
	DisconnectTime             float64   `gorm:"column:DisconnectTime;type:float8"`
	LastDigitSent              float64   `gorm:"column:LastDigitSent;type:float8"`
	ToneDetection              float64   `gorm:"column:ToneDetection;type:float8"`
	PDD                        float64   `gorm:"column:PDD;type:float8"`
	BNumber                    string    `gorm:"column:BNumber;size:30"`
	BNumberDialed              string    `gorm:"column:BNumberDialed;size:30"`
	CallingNumber              string    `gorm:"column:CallingNumber;size:30"`
	Route                      string    `gorm:"column:Route;size:50"`
	CauseCodeID                int       `gorm:"column:CauseCodeId;type:int"`
	CauseLocationID            int       `gorm:"column:CauseLocationId;type:int"`
	Status                     string    `gorm:"column:Status;size:50"`
	AudioURL                   string    `gorm:"column:AudioURL;size:100"`
	AudioFile                  []byte    `gorm:"column:AudioFile;type:bytea"`
	AudioGraph                 []byte    `gorm:"column:AudioGraph;type:bytea"`
	DataLoaded                 bool      `gorm:"column:DataLoaded;default:false"`
	CliDetectedCallingNumber   string    `gorm:"column:CLIDetectedCallingNumber;size:30"`
	CliResult                  string    `gorm:"column:CLIResult;size:250"`
	FasResult                  string    `gorm:"column:FasResult;size:30"`
	VoiceQualityMos            float64   `gorm:"column:voiceQualityMos;type:float8"`
	VoiceQualityNoiceLevel     int       `gorm:"column:voiceQualityNoiceLevel;type:int"`
	VoiceQualitySNR            int       `gorm:"column:voiceQualitySNR;type:int"`
	VoiceQualitySpeechActivity int       `gorm:"column:voiceQualitySpeechActivity;type:int"`
	VoiceQualitySpeechLevel    int       `gorm:"column:voiceQualitySpeechLevel;type:int"`
}

func (PurchOppt) TableName() string {
	return schemaPG + "Purch_Oppt"
}

type PurchOppt struct {
	RequestID              int       `gorm:"column:RequestID;type:int"`
	RequestState           int       `gorm:"column:RequestState;type:int"`
	DestinationID          int       `gorm:"column:DestinationID;type:int"`
	Destination            string    `gorm:"column:Destination;size:50"`
	RouteCarrier           string    `gorm:"column:Route_Carrier;size:20"`
	RouteCarrierID         int       `gorm:"column:Route_CarrierID;type:int"`
	Supplier               string    `gorm:"column:Supplier;size:20"`
	SupplierID             int       `gorm:"column:SupplierID;type:int"`
	TestedByUser           int       `gorm:"column:Tested_by_User;type:int"`
	TestedFrom             time.Time `gorm:"column:Tested_From;type:timestamp"`
	TestedUntil            time.Time `gorm:"column:Tested_Until;type:timestamp"`
	TestASR                float64   `gorm:"column:Test_ASR;type:float8"`
	TestACD                float64   `gorm:"column:Test_ACD;type:float8"`
	TestCalls              int       `gorm:"column:Test_Calls;type:int"`
	TestMinutes            float64   `gorm:"column:Test_Minutes;type:float8"`
	TestType               int       `gorm:"column:Test_Type;type:int"`
	TestComment            string    `gorm:"column:Test_Comment;type:text"`
	TestResult             string    `gorm:"column:Test_Result;size:50"`
	RequestName            string    `gorm:"column:RequestName;size:100"`
	TestingSystemRequestID string    `gorm:"column:TestingSystemRequestID;size:50"`
	LiveTrafficPercentage  int       `gorm:"column:Live_Traffic_Percentage;type:int"`
	ManualTestDetails      []byte    `gorm:"column:Manual_Test_Details;type:bytea"`
	ManualTestDetailsExt   string    `gorm:"column:Manual_Test_Details_Ext;size:50"`
	UserAlerted            bool      `gorm:"column:User_Alerted;default:false"`
	CustomBNumbers         string    `gorm:"column:Custom_BNumbers;type:text"`
	CallingSysRouteID      int       `gorm:"column:CallingSys_RouteID;type:int"`
}

type testtype string

func (t testtype) name() string {
	ts := strings.ToLower(string(t))
	switch {
	case strings.Contains(ts, "cli"):
		return "cli"
	case strings.Contains(ts, "voice"):
		return "voice"
	case strings.Contains(ts, "fas"):
		return "fas"
	}
	return "Unknown Test Type"
}

type foundTest struct {
	RequestByUser          int
	RequestID              int
	TestingSystemRequestID string
	RequestState           int
	RouteCarrier           string
	TestSysRouteID         int
	SupplierID             int
	TestCalls              int
	BNumber                string
	Destination            string
	DestinationID          int
	SystemID               int
	SystemName             string
	TestType               testtype
	TestComment            string
	TimeZone               string
	WebServiceID           int
}

func (PurchStatuses) TableName() string {
	return schemaPG + "Purch_Statuses"
}

type PurchStatuses struct {
	StatusID           int    `gorm:"column:StatusID;type:int;foreignkey:PurchOppt.TestType"`
	StatusName         string `gorm:"column:StatusName;size:30"`
	TestSystem         int    `gorm:"column:TestSystem;type:int"`
	TestSystemCallType string `gorm:"column:TestSystemCallType;size:20"`
}

func (BnumbersHist) TableName() string {
	return schemaPG + "CallingSys_BnumbersHist"
}

type BnumbersHist struct {
	RecordID      int       `gorm:"column:RecordID;type:int;NOT NULL"`
	Bnumber       string    `gorm:"column:Bnumber;size:30"`
	DestinationID int       `gorm:"column:DestinationID;type:int"`
	DateAdd       time.Time `gorm:"column:Date_Add;type:date;default:'now'"`
	UsedInd       int       `gorm:"column:Used_Ind;type:int;default:0"`
}

func (RouteList) TableName() string {
	return schemaPG + "CallingSys_RouteList"
}

type RouteList struct {
	RouteID               int    `gorm:"column:RouteID;type:int"`
	CallingSystemID       int    `gorm:"column:CallingSystemID;type:int;NOT NULL;default:1"`
	CapturaCarrierID      int    `gorm:"column:Captura_CarrierID;type:int"`
	RemoteRouteName       string `gorm:"column:Remote_Route_Name;size:100;NOT NULL"`
	RemoteRouteID         int    `gorm:"column:Remote_Route_ID;type:int"`
	NetsenseIntAccessCode string `gorm:"column:Netsense_IntAccessCode;size:10"`
	Dialer                string `gorm:"column:Dialer;size:50"`
	Prefix                string `gorm:"column:Prefix;size:20"`
	Codec                 string `gorm:"column:Codec;size:10"`
}
