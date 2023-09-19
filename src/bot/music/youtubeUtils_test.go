package music

import (
	"testing"
	"time"

	"github.com/CarlFlo/dootBot/src/test"
)

func TestYoutubeTimeToDuration(t *testing.T) {

	input := "PT1H24M47S"
	duration := youtubeTimeToDuration(input)

	answer := time.Hour*1 + time.Minute*24 + time.Second*47

	test.Validate(t, duration, answer, "The durations should match")
}
