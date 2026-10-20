package blmkc

import "time"

type Summary struct {
	Total          int
	Wanted         int
	Identified     int
	Apprehended    int
	Resolved       int
	PoliceReported int
	PendingReview  int
	LossAmount     float64
}

type StatusCount struct {
	Status string
	Count  int
}

type Case struct {
	ID                 uint64
	CaseNumber         string
	StoreName          string
	Title              string
	OccurredAt         string
	CaseStatus         string
	VerificationStatus string
	IncidentType       string
	PrimarySubject     string
	ItemCount          int
	EvidenceCount      int
}

type DashboardData struct {
	Summary      Summary
	RecentCases  []Case
	StatusCounts []StatusCount
}

type Filter struct {
	Search             string
	StoreID            uint64
	CaseStatus         string
	VerificationStatus string
	StartDate          string
	EndDate            string
	Page               int
	PageSize           int
}

type Option struct {
	ID   uint64
	Code string
	Name string
}

type CaseListData struct {
	Cases      []Case
	Stores     []Option
	Filter     Filter
	Total      int
	TotalPages int
	PrevPage   int
	NextPage   int
}

type WantedData struct {
	Cases           []Case
	Search          string
	Status          string
	WantedCount     int
	IdentifiedCount int
}

type Identity struct {
	ID             uint64
	SubjectNumber  string
	FullName       string
	Aliases        string
	IdentityStatus string
	CaseCount      int
	StoreCount     int
	LastCaseDate   string
	LastCaseStatus string
}

type IdentityData struct {
	Subjects []Identity
	Search   string
	Total    int
}

type VerificationData struct {
	Cases     []Case
	Status    string
	Search    string
	Submitted int
	Revision  int
	Verified  int
	Rejected  int
}

type FormOption struct {
	ID      uint64
	StoreID uint64
	Code    string
	Name    string
}

type CaseInput struct {
	StoreID        uint64
	IncidentTypeID *uint64
	AreaID         *uint64
	CameraID       *uint64
	OccurredAt     string
	Title          string
	Chronology     string
	LossAmount     *float64
	SubjectName    string
	InitialStatus  string
}

type CaseFormData struct {
	Case                   CaseInput
	CaseID                 uint64
	CaseNumber             string
	SelectedIncidentTypeID uint64
	SelectedAreaID         uint64
	SelectedCameraID       uint64
	Stores                 []FormOption
	IncidentTypes          []FormOption
	Areas                  []FormOption
	Cameras                []FormOption
	IsEdit                 bool
}

type CaseItemDetail struct {
	Name, Brand, Quantity, Value, Status, Notes string
}

type VehicleDetail struct {
	Type, Brand, Model, Color, Plate, Notes string
}

type EvidenceDetail struct {
	ID                                                    uint64
	FileSize                                              uint64
	Category, Filename, Description, CreatedAt, SizeLabel string
	MIMEType, VerificationStatus, VerificationNotes       string
}

type EvidenceFile struct {
	ID, CaseID, StoreID, FileSize             uint64
	Category, Filename, StoragePath, MIMEType string
}

type EvidenceInput struct {
	CaseID, SubjectID, UserID, FileSize                    uint64
	Category, Filename, StoragePath, MIMEType, Description string
	FileHash                                               []byte
	CapturedAt                                             *time.Time
}

type ResolutionDetail struct {
	Type, Date, Summary, PoliceNumber, PoliceStation, Compensation string
}

type EventDetail struct {
	Actor, EventType, FromStatus, ToStatus, Notes, CreatedAt string
}

type CaseDetail struct {
	ID                    uint64
	CaseNumber            string
	StoreID               uint64
	StoreName             string
	Title                 string
	Chronology            string
	OccurredAt            string
	ReportedAt            string
	CaseStatus            string
	VerificationStatus    string
	LegalStatus           string
	LossAmount            string
	IncidentType          string
	AreaName              string
	CameraName            string
	ReporterName          string
	ReviewerName          string
	SubjectID             uint64
	SubjectNumber         string
	SubjectName           string
	SubjectIdentityStatus string
	SubjectAliases        string
	CanAdvance            bool
	NextStatus            string
	Items                 []CaseItemDetail
	Vehicles              []VehicleDetail
	Evidence              []EvidenceDetail
	Resolutions           []ResolutionDetail
	Events                []EventDetail
}
