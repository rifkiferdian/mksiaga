package area

type Area struct {
	ID          uint64
	StoreID     uint64
	StoreCode   string
	StoreName   string
	Code        string
	Name        string
	IsPriority  bool
	SortOrder   uint16
	Status      string
	CameraCount int
}

type StoreOption struct {
	ID   uint64
	Code string
	Name string
}

type Input struct {
	StoreID    uint64
	Code       string
	Name       string
	IsPriority bool
	SortOrder  uint16
	Status     string
}
