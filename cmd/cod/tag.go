package main

import (
	"strings"
)

func tagSearchCast(tag string) string {
	// `bson:"pageId" json:"pageId"`
	// Example: `cod.cast:"uint64"`

	split := strings.Split(tag, " ")

	// debugPrintln("AAA")
	// debugPrintln(split)
	for _, s := range split {
		s = strings.TrimSpace(s)
		// debugPrintln(s)

		valQuoted, ok := strings.CutPrefix(s, "`cod.cast:")
		if !ok { continue }
		// debugPrintln(valQuoted)

		val := strings.Trim(valQuoted, "\"`")
		// debugPrintln(val)
		return val
	}
	return ""
}


func tagSearchSkip(tag string) string {
	// `bson:"pageId" json:"pageId"`
	// Example: `cod.skip:"equality"`

	split := strings.Split(tag, " ")

	// debugPrintln("AAA")
	// debugPrintln(split)
	for _, s := range split {
		s = strings.TrimSpace(s)
		// debugPrintln(s)

		valQuoted, ok := strings.CutPrefix(s, "`cod.skip:")
		if !ok { continue }
		// debugPrintln(valQuoted)

		val := strings.Trim(valQuoted, "\"`")
		// debugPrintln(val)
		return val
	}
	return ""
}

func shouldSkipEquality(skipString string) bool {
	return strings.Contains(skipString, "equality")
}
func shouldSkipSerdes(tag string) bool {
	skip := tagSearchSkip(tag)
	return strings.Contains(skip, "serdes")
}
