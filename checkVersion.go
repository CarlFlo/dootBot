package main

import (
	"io/ioutil"
	"net/http"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"
)

/*
	Return true or false if the version is up to date
	Return version on system
	Return version on github
	return error
*/
func botVersonHandler() (bool, string, string, error) {

	current := currentVersion()
	githubVersion, err := githubVersion()

	if err != nil {
		return false, current, "", err
	}

	upToDate := current == githubVersion

	return upToDate, current, githubVersion, nil
}

func currentVersion() string {

	// read file from directory
	file, err := ioutil.ReadFile("./version")
	if err != nil {
		malm.Fatal("%s", err)
	}

	return string(file)
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

	return string(body), nil
}
