package skedda

// SpaceList is an array of *Space
type SpaceList []*Space

// FindByID returns space with matching ID
func (l SpaceList) FindByID(id int) *Space {
	for _, v := range l {
		if v.ID == id {
			return v
		}
	}
	return nil
}

// Map calls `mapper` on each space and returns the result
func (l SpaceList) Map(mapper func(i int, v Space) string) []string {
	result := []string{}
	for i, v := range l {
		result = append(result, mapper(i, *v))
	}

	return result
}

func (l SpaceList) Len() int      { return len(l) }
func (l SpaceList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l SpaceList) Less(i, j int) bool {
	if l[i].VenueID == l[j].VenueID {
		return l[i].Name < l[j].Name
	}
	return l[i].VenueID < l[j].VenueID
}
