package skedda

import (
	"strings"
	"time"
)

// DateTime with custom format: 2006-01-02T15:04:05
type DateTime struct {
	time.Time
}

// UnmarshalJSON decodes datetime
func (dt *DateTime) UnmarshalJSON(input []byte) error {
	strInput := string(input)
	strInput = strings.Trim(strInput, `"`)
	newTime, err := time.Parse("2006-01-02T15:04:05", strInput)
	if err != nil {
		return err
	}

	dt.Time = newTime
	return nil
}
