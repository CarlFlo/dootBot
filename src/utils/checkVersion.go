package utils

import (
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/malm"
)

func CheckVersion(currentVersion string) {
	// Handles checking if there is an update available for the bot
	upToDate, githubVersion, err := botVersonHandler(currentVersion)
	if err != nil {
		malm.Error("%s", err)
	}

	if upToDate {
		malm.Debug("Version %s", currentVersion)
	} else {
		malm.Info("New version available at '%s'! New version: '%s'; Your version: '%s'",
			config.CONFIG.BotInfo.DepositURL,
			githubVersion,
			currentVersion)
	}
}

// Return true or false if the version is up to date
// Return version on system
// Return version on github
// return error
func botVersonHandler(current string) (bool, string, error) {

	githubVersion, err := githubVersion()

	if err != nil {
		return false, "", err
	}

	upToDate := current == githubVersion

	return upToDate, githubVersion, nil
}

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
