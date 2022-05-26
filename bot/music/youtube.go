package music

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"

	"github.com/bwmarrin/discordgo"
)

type videoResponse struct {
	Formats []struct {
		Url string `json:"url"`
	} `json:"formats"`
}

// youtube-dl --skip-download --print-json --flat-playlist J-innQH71As

func youtubeDL(m *discordgo.MessageCreate, videoID string) (Song, error) {

	cmd := exec.Command("youtube-dl", "--skip-download", "--print-json", "--flat-playlist", videoID)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return Song{}, err
	}

	var videoRes videoResponse
	err = json.NewDecoder(&out).Decode(&videoRes)
	if err != nil {
		log.Println("Could not decode the video")
		return Song{}, err
	}

	// Query the name of the song from youtube using their endpoint

	song := Song{
		ChannelID:  m.ChannelID,
		Title:      "Placeholder title",
		YoutubeURL: videoID,
		User:       m.Author.Username,
	}

	return song, nil
}
