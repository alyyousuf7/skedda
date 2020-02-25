package skedda

import "fmt"

// Booking of a space in Skedda
type Booking struct {
	ID             int
	Title          string
	StartTime      DateTime `json:"start"`
	EndTime        DateTime `json:"end"`
	RecurrenceRule RuleSet
	SpaceIDs       []int `json:"spaces"`
	VenueID        int   `json:"venue"`
}

func (b Booking) String() string {
	dateTimeFormat := "2006-01-02 03:04pm"
	timeFormat := "03:04pm"

	if b.Title == "" {
		b.Title = "[Unknown]"
	}

	if len(b.RecurrenceRule.All()) > 0 {
		return fmt.Sprintf("%s -- %s - %s (Recurring)", b.Title, b.StartTime.Format(timeFormat), b.EndTime.Format(timeFormat))
	}

	return fmt.Sprintf("%s -- %s - %s", b.Title, b.StartTime.Format(dateTimeFormat), b.EndTime.Format(timeFormat))
}
