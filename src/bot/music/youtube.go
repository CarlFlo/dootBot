package music

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
)

type videoResponse struct {
	Formats []struct {
		Url string `json:"url"`
	} `json:"formats"`
}

// youtube-dl --skip-download --print-json --flat-playlist J-innQH71As

func execYoutubeDL(song *Song) error {

	cmd := exec.Command("youtube-dl", "--skip-download", "--print-json", "--flat-playlist", song.YoutubeURL)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return err
	}

	var videoRes videoResponse
	err = json.NewDecoder(&out).Decode(&videoRes)
	if err != nil {
		log.Println("Could not decode the video")
		return err
	}

	// The URL directely to the audio
	song.StreamURL = videoRes.Formats[0].Url
	return nil
}
