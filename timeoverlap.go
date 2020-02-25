package skedda

import (
	"time"
)

// TimeOverlaps returns true if t1 range is overlapping or being overlapped by t2 range
func TimeOverlaps(t1Start, t1End, t2Start, t2End time.Time) bool {
	// wrong inputs
	if t1Start.After(t1End) || t2Start.After(t2End) {
		return false
	}

	// subset
	if (t2Start.After(t1Start) && t2Start.Before(t1End)) || (t2End.After(t1Start) && t2End.Before(t1End)) {
		return true
	}

	// superset
	if (t2Start.Before(t1Start) && t2End.After(t1Start)) || (t2Start.Before(t1End) && t2End.After(t1End)) {
		return true
	}

	// equal
	if t2Start.Equal(t1Start) && t2End.Equal(t1End) {
		return true
	}

	return false
}
