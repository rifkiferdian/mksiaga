package checklist

type Checklist struct {
	ID                  uint64
	Code                string
	Phase               string
	Name                string
	RequiresCameraCount bool
	SortOrder           uint16
	Status              string
}

type Input struct {
	Code                string
	Phase               string
	Name                string
	RequiresCameraCount bool
	SortOrder           uint16
	Status              string
}
