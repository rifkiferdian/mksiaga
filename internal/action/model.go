package action

type Action struct {
	ID                  uint64
	Code                string
	Name                string
	RequiresDescription bool
	SortOrder           uint16
	Status              string
}

type Input struct {
	Code                string
	Name                string
	RequiresDescription bool
	SortOrder           uint16
	Status              string
}
