package store

type Store struct {
	ID        uint64
	Code      string
	Name      string
	Address   string
	Phone     string
	Timezone  string
	Status    string
	UserCount int
}

type Input struct {
	Code     string
	Name     string
	Address  string
	Phone    string
	Timezone string
	Status   string
}
