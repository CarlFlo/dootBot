package utils

import (
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/CarlFlo/DiscordMoneyBot/src/config"
)

//	Return true or false if the version is up to date
//	Return version on system
//	Return version on github
//	return error
func BotVersonHandler(current string) (bool, string, error) {

	githubVersion, err := githubVersion()

	if err != nil {
		return false, "", err
	}

	upToDate := current == githubVersion

	return upToDate, githubVersion, nil
}

/*
func currentVersion() string {

	// read file from directory
	file, err := ioutil.ReadFile("./version")
	if err != nil {
		malm.Fatal("%s", err)
	}

	return string(file)
}
*/

// Returns the online version or the error
func githubVersion() (string, error) {

	// get URL
	resp, err := http.Get(config.CONFIG.BotInfo.VersionURL)
	if err != nil {
		return "", err
	}

	// read response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// regex to find version
	pattern := regexp.MustCompile(`\d+-\d+-\d+`)
	version := pattern.FindString(string(body))

	return version, nil
}
