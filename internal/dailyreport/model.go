package dailyreport

type Option struct {
	ID   uint64
	Name string
	Code string
}

type ChecklistOption struct {
	ID                  uint64
	Name                string
	Phase               string
	RequiresCameraCount bool
}

type Report struct {
	ID          uint64
	ReportDate  string
	ShiftName   string
	OfficerName string
	Status      string
	CreatedAt   string
}

type ChecklistAnswer struct {
	ChecklistID uint64
	Answer      string
	Active      *uint16
	Total       *uint16
	Notes       string
}

type MonitoringLog struct {
	Time      string
	AreaID    *uint64
	Condition string
	Finding   string
	Action    string
}

type Incident struct {
	Time             string
	AreaID           *uint64
	CameraID         *uint64
	IncidentTypeID   *uint64
	OtherType        string
	Description      string
	Gender           string
	Age              string
	Clothing         string
	Features         string
	ResponseMinutes  *uint16
	Status           string
	ActionIDs        []uint64
	OtherActionNotes string
}

type Input struct {
	StoreID          uint64
	OfficerID        uint64
	SupervisorID     *uint64
	ShiftID          uint64
	ReportDate       string
	WorkStart        string
	WorkEnd          string
	Location         string
	StartNotes       string
	EndNotes         string
	HandoverToUserID *uint64
	HandoverAt       string
	HandoverNotes    string
	HandoverFromSign string
	HandoverToSign   string
	Status           string
	Checklists       []ChecklistAnswer
	Monitoring       []MonitoringLog
	Incidents        []Incident
}

type PageData struct {
	Reports       []Report
	Shifts        []Option
	Areas         []Option
	PriorityAreas []Option
	Cameras       []Option
	Users         []Option
	Checklists    []ChecklistOption
	IncidentTypes []Option
	Actions       []Option
}
