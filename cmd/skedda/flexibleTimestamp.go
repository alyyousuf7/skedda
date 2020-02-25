package main

import (
	"fmt"
	"time"
)

// FlexibleTimestamp supports time with flexible layout
type FlexibleTimestamp struct {
	Layouts         []string
	Time            *time.Time
	UsedLayoutIndex int
}

// Set parses the string value to timestamp
func (v *FlexibleTimestamp) Set(value string) error {
	if len(v.Layouts) == 0 {
		return fmt.Errorf("no layouts provided")
	}

	for i, layout := range v.Layouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			v.Time = &t
			v.UsedLayoutIndex = i
			return nil
		}
	}

	return fmt.Errorf("no time format matched %s", value)
}

func (v FlexibleTimestamp) String() string {
	if v.Time == nil {
		return ""
	}

	if len(v.Layouts) == 0 {
		panic(fmt.Errorf("no layouts provided"))
	}

	return v.Time.Format(v.Layouts[v.UsedLayoutIndex])
}

// Get returns the flag structure
func (v *FlexibleTimestamp) Get() interface{} {
	return v.Time
}
