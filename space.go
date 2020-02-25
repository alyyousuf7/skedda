package skedda

// Space of a Venue in Skedda
type Space struct {
	ID      int
	Name    string
	VenueID int `json:"venue"`
}

func (s Space) String() string {
	return s.Name
}
