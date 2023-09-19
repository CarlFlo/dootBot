package music

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/CarlFlo/malm"
)

func execYoutubeDL(song *Song) error {
	// Stream URL it returns should be valid for 6 hours
	cmd := exec.Command("yt-dlp", song.YoutubeVideoID, "--skip-download", "--get-url", "--flat-playlist")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return err
	}

	lines := strings.Split(out.String(), "\n")

	if len(lines) >= 2 {
		song.StreamURL = lines[1]
	} else {
		malm.Error("youtube DL - There are not enough lines of output.")
	}

	return nil
}

/*
type videoResponse struct {
	Formats []struct {
		Url string `json:"url"`
	} `json:"formats"`
}

func execYoutubeDLOld(song *Song) error {

	cmd := exec.Command("yt-dlp", song.YoutubeVideoID, "--skip-download", "--print-json", "--flat-playlist")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return err
	}

	var videoRes videoResponse
	err = json.NewDecoder(&out).Decode(&videoRes)
	if err != nil {
		return err
	}

	// The URL directly to the audio. Expires after 6 hours
	// yt-dlp uses index 3. Youtube-dl uses index 0
	song.StreamURL = videoRes.Formats[3].Url
	return nil
}
*/
