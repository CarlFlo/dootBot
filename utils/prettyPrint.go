package utils

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type number interface {
	uint64 | int
}

// HumanReadableNumber - turns 100000 into 100,000. Making it much easier to read.
// Accepts both uint64 and int
func HumanReadableNumber[T number](number T) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", number)
}
