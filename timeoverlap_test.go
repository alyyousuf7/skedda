package skedda_test

import (
	"testing"
	"time"

	"github.com/alyyousuf7/skedda"
)

func TestTimeOverlaps(t *testing.T) {
	start := time.Date(1991, time.March, 7, 8, 0, 0, 0, time.UTC)
	end := time.Date(1991, time.March, 7, 12, 0, 0, 0, time.UTC)

	// t1       t2
	//          |
	//          | |
	//    _  _  _ | | _  _  _  _  _ |_  |
	// |            | |           | | | |
	// |              | |         | | | |
	// |                | |       | | | |
	// |  _  _  _  _  _  _| |  _  | |_| |
	//                      | |     | |
	//                        | |
	//                          |
	// Expected F F T T T T T F F T T T T
	cases := []struct {
		start    time.Time
		end      time.Time
		expected bool
	}{
		{
			start:    time.Date(1991, time.March, 7, 5, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 7, 0, 0, 0, time.UTC),
			expected: false,
		}, {
			start:    time.Date(1991, time.March, 7, 6, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 8, 0, 0, 0, time.UTC),
			expected: false,
		}, {
			start:    time.Date(1991, time.March, 7, 7, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 9, 0, 0, 0, time.UTC),
			expected: true,
		}, {
			start:    time.Date(1991, time.March, 7, 8, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 10, 0, 0, 0, time.UTC),
			expected: true,
		}, {
			start:    time.Date(1991, time.March, 7, 9, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 11, 0, 0, 0, time.UTC),
			expected: true,
		}, {
			start:    time.Date(1991, time.March, 7, 10, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 12, 0, 0, 0, time.UTC),
			expected: true,
		}, {
			start:    time.Date(1991, time.March, 7, 11, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 13, 0, 0, 0, time.UTC),
			expected: true,
		}, {
			start:    time.Date(1991, time.March, 7, 12, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 14, 0, 0, 0, time.UTC),
			expected: false,
		}, {
			start:    time.Date(1991, time.March, 7, 13, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 15, 0, 0, 0, time.UTC),
			expected: false,
		}, {
			start:    time.Date(1991, time.March, 7, 8, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 12, 0, 0, 0, time.UTC),
			expected: true,
		}, {
			start:    time.Date(1991, time.March, 7, 7, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 13, 0, 0, 0, time.UTC),
			expected: true,
		}, {
			start:    time.Date(1991, time.March, 7, 8, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 13, 0, 0, 0, time.UTC),
			expected: true,
		}, {
			start:    time.Date(1991, time.March, 7, 7, 0, 0, 0, time.UTC),
			end:      time.Date(1991, time.March, 7, 12, 0, 0, 0, time.UTC),
			expected: true,
		},
	}

	for i, c := range cases {
		result := skedda.TimeOverlaps(start, end, c.start, c.end)
		if c.expected != result {
			t.Errorf("Expected %v but got %v: Test case %d: %q - %q", c.expected, result, i+1, c.start, c.end)
		}
	}
}
