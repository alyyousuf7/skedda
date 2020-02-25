package skedda

// VenueList is an array of *Venue
type VenueList []*Venue

// FindByID returns venue with matching ID
func (l VenueList) FindByID(id int) *Venue {
	for _, v := range l {
		if v.ID == id {
			return v
		}
	}
	return nil
}

// Map calls `mapper` on each venue and returns the result
func (l VenueList) Map(mapper func(i int, v Venue) string) []string {
	result := []string{}
	for i, v := range l {
		result = append(result, mapper(i, *v))
	}

	return result
}
