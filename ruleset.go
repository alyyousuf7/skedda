package skedda

import (
	"regexp"
	"strings"

	"github.com/teambition/rrule-go"
)

// RuleSet is iCal Recurrence Rule Set
type RuleSet struct {
	rrule.Set
	ForceValid bool
}

// UnmarshalJSON decodes the iCal Recurrence Rule Set
func (r *RuleSet) UnmarshalJSON(input []byte) error {
	if string(input) == "null" {
		return nil
	}

	// Remove DTEND
	if !r.ForceValid {
		re := regexp.MustCompile(`DTEND:([0-9]+T[0-9]+Z)`)
		input = re.ReplaceAll(input, []byte(""))
	}

	strInput := string(input)
	strInput = strings.ReplaceAll(strInput, `\r`, "")
	strInput = strings.ReplaceAll(strInput, `\n\n`, "\n")
	strInput = strings.ReplaceAll(strInput, `\n`, "\n")
	strInput = strings.Trim(strInput, `"`)

	set, err := rrule.StrToRRuleSet(strInput)
	if err != nil {
		return err
	}

	r.Set = *set
	return nil
}
