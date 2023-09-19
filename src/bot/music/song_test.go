package music

import (
	"testing"
	"time"

	"github.com/CarlFlo/dootBot/src/test"
)

func TestGetDuration(t *testing.T) {

	var output string
	song := Song{
		Duration: time.Hour*1 + time.Minute*6 + time.Second*26,
	}

	//Test 1
	output = song.GetDuration()
	test.Validate(t, output, "1h 6m 26s", "")

	//Test 2
	song = Song{
		Duration: time.Hour*1 + time.Second*26,
	}
	output = song.GetDuration()
	test.Validate(t, output, "1h 0m 26s", "")

	//Test 3
	song = Song{
		Duration: time.Second * 45,
	}
	output = song.GetDuration()
	test.Validate(t, output, "45s", "")
}
