package music

import (
	"testing"

	"github.com/CarlFlo/DiscordMoneyBot/src/test"
)

func TestFormatYoutubeDuration(t *testing.T) {

	r1 := formatYoutubeDuration("PT1H24M47S")
	test.Validate(t, r1, "1h 24m 47s", "")

	r2 := formatYoutubeDuration("PT12M12S")
	test.Validate(t, r2, "12m 12s", "")

	r3 := formatYoutubeDuration("PT60S")
	test.Validate(t, r3, "60s", "")

	r4 := formatYoutubeDuration("PT")
	test.Validate(t, r4, "", "")

}
