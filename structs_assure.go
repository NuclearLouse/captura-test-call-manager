// structs_assure.go
//
// The file contains the structures that are needed
// to process JSON responses from the "Assure" system
// and contains the structures that are needed to work with "Assure" tables.
//
package main

func (assureAPI) TableName() string {
	return schemaPG + "CallingSys_API_Assure"
}

type assureAPI struct {
	SystemName        string `gorm:"size:50;foreignkey:CallingSys_Settings.SystemName"`
	SystemID          int    `gorm:"type:int"`
	URL               string `gorm:"size:100"`
	User              string `gorm:"size:100"`
	Pass              string `gorm:"size:100"`
	StatusTests       string `gorm:"size:100"`
	TestResults       string `gorm:"size:100"`
	QueryResults      string `gorm:"size:100"`
	Routes            string `gorm:"size:25"`
	Destinations      string `gorm:"size:25"`
	Nodes             string `gorm:"size:25"`
	NewTestGet        string `gorm:"size:25"`
	NodesCapabilities string `gorm:"size:25"`
	Version           string `gorm:"size:25"`
}

//-----------------------------------------------------------------------------
// *******************Structs for initiated new tests*******************
//-----------------------------------------------------------------------------
type testSetDestination struct {
	NoOfExecutions int
	TestSetItems   []batchDestination
}

type batchDestination struct {
	TestTypeName  string
	RouteID       int
	DestinationID int
}

type testSetBnumbers struct {
	TestSetItems []batchBnumbers
}

type batchBnumbers struct {
	TestTypeName string
	RouteID      int
	PhoneNumber  int64
}

//-----------------------------------------------------------------------------
// ****************Structs for obtain tests status and results****************
//-----------------------------------------------------------------------------
type testStatusAssure struct {
	TestBatchID    int    `json:"TestBatchID"`
	StatusID       int    `json:"StatusID"`
	Status         string `json:"Status"`
	IsDone         bool   `json:"IsDone"`
	Created        string `json:"Created"` //time.Time
	TestBatchItems []struct {
		IsDone   bool   `json:"IsDone"`
		Status   string `json:"Status"`
		StatusID int    `json:"StatusID"`
	} `json:"TestBatchItems"`
}

type testResultAssure struct {
	QueryResult1 []batchResult
}

type batchResult struct {
	CallResultID            int
	ANumber                 string `json:"A number"`
	ANumberIsEncrypted      bool   `json:"A_NumberIsEncrypted"`
	BNumberIsEncrypted      bool   `json:"B_NumberIsEncrypted"`
	Result                  string
	TestType                string `json:"Test Type"`
	BNetwork                string `json:"B Network"`
	BTestNode               string `json:"B Test Node"`
	Bnumber                 string `json:"B number (first 6 digits)"`
	Route                   string
	TestStartTime           string `json:"Test Start Time"` //time.Time
	Alert                   interface{}
	Connect                 interface{}
	CLI                     interface{} //int
	VQA                     interface{} `json:"VQ A"`
	VQB                     interface{} `json:"VQ B"`
	FAS                     interface{} //int
	CLIDelivered            string      `json:"CLI Delivered"`
	MOSA                    float64     `json:"MOS A"`
	MOSB                    float64     `json:"MOS B"`
	PGRD                    float64
	PGAD                    float64
	CallDuration            float64     `json:"Call Duration"`
	DisconnectTime          float64     `json:"Disconnect Time"`
	BAlertTime              float64     `json:"B Alert Time"`
	BCallDuration           float64     `json:"B Call Duration"`
	BConnectTime            float64     `json:"B Connect Time"`
	BDisconnectTime         float64     `json:"B Disconnect Time"`
	ReleaseCause            string      `json:"Release Cause"`
	ReleaseLocation         string      `json:"Release Location"`
	SpeechFirstDetectedTime interface{} `json:"Speech First Detected Time"` //time.Time
	SpeechLastDetectedTime  interface{} `json:"Speech Last Detected Time"`  //time.Time
	APartyAudio             string      `json:"A Party Audio"`
}
