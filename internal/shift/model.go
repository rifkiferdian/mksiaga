package shift

type Shift struct {
	ID        uint64
	Code      string
	Name      string
	StartTime string
	EndTime   string
	SortOrder uint16
	Status    string
}

type Input struct {
	Code      string
	Name      string
	StartTime string
	EndTime   string
	SortOrder uint16
	Status    string
}
