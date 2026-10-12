package camera

type Camera struct {
	ID          uint64
	StoreID     uint64
	StoreCode   string
	StoreName   string
	AreaID      uint64
	AreaCode    string
	AreaName    string
	Code        string
	Name        string
	Description string
	Status      string
}

type StoreOption struct {
	ID   uint64
	Code string
	Name string
}

type AreaOption struct {
	ID      uint64
	StoreID uint64
	Code    string
	Name    string
}

type Input struct {
	StoreID     uint64
	AreaID      uint64
	Code        string
	Name        string
	Description string
	Status      string
}
