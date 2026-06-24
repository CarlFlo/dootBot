package music

import (
	"testing"
	"time"

	"github.com/CarlFlo/dootBot/src/test"
)

func TestFormatDuration(t *testing.T) {
	test.Validate(t, formatDuration(0), "0:00", "")
	test.Validate(t, formatDuration(45*time.Second), "0:45", "")
	test.Validate(t, formatDuration(2*time.Minute+38*time.Second), "2:38", "")
	test.Validate(t, formatDuration(time.Hour+6*time.Minute+26*time.Second), "1:06:26", "")
}
