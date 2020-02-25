package skedda

// Venue in Skedda
type Venue struct {
	ID     int
	Name   string
	Domain string `json:"subdomain"`
}

func (v Venue) String() string {
	return v.Name
}
