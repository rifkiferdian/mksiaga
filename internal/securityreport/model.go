package securityreport

type Option struct {
	ID        uint64
	Name      string
	Code      string
	StartTime string
	EndTime   string
}

type ListItem struct {
	ID, MemberCount, LogCount                uint64
	ReportDate, StoreName, ShiftName         string
	RepresentativeName, Status, HighestLevel string
	WorkStartedAt, WorkEndedAt               string
}

type Member struct {
	ID       uint64
	UserID   uint64
	Name     string
	Role     string
	JoinedAt string
}

type HourlyLog struct {
	ID                                                uint64
	ReportedAt, Level, AreaName, LocationNotes        string
	Description, ActionTaken, FollowUpStatus, Creator string
}

type Detail struct {
	ID, RepresentativeID                                     uint64
	ReportDate, StoreName, ShiftName, RepresentativeName     string
	SupervisorName, WorkStartedAt, WorkEndedAt, OpeningNotes string
	ClosingSummary, Status, HighestLevel                     string
	NormalCount, WaspadaCount, UrgentCount                   int
	Members                                                  []Member
	Logs                                                     []HourlyLog
}

type PageData struct {
	Reports     []ListItem
	Shifts      []Option
	Users       []Option
	Supervisors []Option
	Areas       []Option
}

type CreateInput struct {
	StoreID, ShiftID, RepresentativeID, CreatorID uint64
	SupervisorID                                  *uint64
	ReportDate, OpeningNotes                      string
	MemberIDs                                     []uint64
}

type HourlyLogInput struct {
	ReportID, CreatorID uint64
	AreaID              *uint64
	Level               string
	LocationNotes       string
	Description         string
	ActionTaken         string
	FollowUpStatus      string
}
