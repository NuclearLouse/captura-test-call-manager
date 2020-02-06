package main

//-----------------------------------------------------------------------------
//*******************Block of Assure syncro structures*******************
//-----------------------------------------------------------------------------
type destinations struct {
	QueryResult1 []assureDestination
}

func (assureDestination) TableName() string {
	return schemaPG + "CallingSys_assure_destinations"
}

type assureDestination struct {
	CountryID                 int         `json:"CountryID" gorm:"type:int"`
	CountryName               string      `json:"CountryName" gorm:"type:varchar(100)"`
	Created                   string      `json:"Created" gorm:"type:varchar(100)"`
	CreatedBy                 int         `json:"CreatedBy" gorm:"type:int"`
	DestinationCategoryID     int         `json:"DestinationCategoryID" gorm:"type:int"`
	DestinationCategoryName   string      `json:"DestinationCategoryName" gorm:"type:varchar(100)"`
	DestinationExtID          interface{} `json:"DestinationExtID" gorm:"type:varchar(100)"`
	DestinationID             int         `json:"DestinationID" gorm:"type:int"`
	DestinationImportanceID   int         `json:"DestinationImportanceID"`
	DestinationImportanceName string      `json:"DestinationImportanceName" gorm:"type:varchar(100)"`
	Modified                  string      `json:"Modified" gorm:"type:varchar(100)"`
	ModifiedBy                int         `json:"ModifiedBy" gorm:"type:int"`
	Name                      string      `json:"Name" gorm:"type:varchar(100)"`
	PoPID                     int         `json:"PoPID" gorm:"type:int"`
	ShortName                 string      `json:"ShortName" gorm:"type:varchar(100)"`
}

type routes struct {
	QueryResult1 []assureRoute
}

func (assureRoute) TableName() string {
	return schemaPG + "CallingSys_assure_routes"
}

type assureRoute struct {
	Active                 bool        `json:"Active" gorm:"type:boolean"`
	CallParameter          interface{} `json:"CallParameter" gorm:"type:varchar(100)"`
	Carrier                string      `json:"Carrier" gorm:"type:varchar(100)"`
	CarrierID              interface{} `json:"CarrierID" gorm:"type:varchar(100)"`
	Channel                int         `json:"Channel" gorm:"type:int"`
	ChannelPoolID          int         `json:"ChannelPoolID" gorm:"type:int"`
	ChannelPoolName        string      `json:"ChannelPoolName" gorm:"type:varchar(100)"`
	Created                string      `json:"Created" gorm:"type:varchar(100)"`
	CreatedBy              int         `json:"CreatedBy" gorm:"type:int"`
	Description            interface{} `json:"Description" gorm:"type:varchar(100)"`
	DialerName             string      `json:"DialerName" gorm:"type:varchar(100)"`
	ExtRouteID             interface{} `json:"ExtRouteID" gorm:"type:varchar(100)"`
	IsMainProduct          interface{} `json:"IsMainProduct" gorm:"type:varchar(100)"`
	Modified               string      `json:"Modified" gorm:"type:varchar(100)"`
	ModifiedBy             int         `json:"ModifiedBy" gorm:"type:int"`
	Name                   string      `json:"Name" gorm:"type:varchar(100)"`
	NumberManipulationRule interface{} `json:"NumberManipulationRule" gorm:"type:varchar(100)"`
	PoPID                  int         `json:"PoPID" gorm:"type:int"`
	Prefix                 string      `json:"Prefix" gorm:"type:varchar(100)"`
	RouteClass             string      `json:"RouteClass" gorm:"type:varchar(100)"`
	RouteID                int         `json:"RouteID" gorm:"type:int"`
	RouteImportanceID      int         `json:"RouteImportanceID" gorm:"type:int"`
	RouteImportanceName    string      `json:"RouteImportanceName" gorm:"type:varchar(100)"`
	RouteTypeID            interface{} `json:"RouteTypeID" gorm:"type:varchar(100)"`
	RouteTypeName          interface{} `json:"RouteTypeName" gorm:"type:varchar(100)"`
	ShortName              string      `json:"ShortName" gorm:"type:varchar(100)"`
	SwitchID               int         `json:"SwitchID" gorm:"type:int"`
	SwitchName             string      `json:"SwitchName" gorm:"type:varchar(100)"`
}
