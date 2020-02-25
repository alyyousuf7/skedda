package main

import (
	"fmt"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Matcher for venues and spaces
type Matcher struct {
	list  []fmt.Stringer
	words []string
}

// NewMatcher constructs Matcher with items
func NewMatcher(list []fmt.Stringer) *Matcher {
	words := []string{}

	for _, item := range list {
		words = append(words, item.String())
	}

	return &Matcher{list, words}
}

// Match the string and return closest item
func (m *Matcher) Match(str string) []fmt.Stringer {
	matches := fuzzy.FindNormalizedFold(str, m.words)

	if len(matches) == 0 {
		return nil
	}

	result := []fmt.Stringer{}
	for _, str := range matches {
		for _, item := range m.list {
			if item.String() == str {
				result = append(result, item)
				break
			}
		}
	}

	return result
}

// MatchMultiple matches each string and returns closest items
func (m *Matcher) MatchMultiple(strv []string) []fmt.Stringer {
	result := []fmt.Stringer{}
	for _, str := range strv {
		match := m.Match(str)
		result = append(result, match...)
	}

	return uniqueStringer(result)
}

func uniqueStringer(stringerSlice []fmt.Stringer) []fmt.Stringer {
	keys := make(map[fmt.Stringer]bool)
	list := []fmt.Stringer{}
	for _, entry := range stringerSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
