package main

import (
	"crypto/rand"
	"math/big"

	"github.com/charmbracelet/hotdiva2000"
	"github.com/google/uuid"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

// LETTERS is a list of lowercase letters used for generating random strings.
var LETTERS = []rune("abcdefghijklmnopqrstuvwxyz")

func randomInt(maxInt int) int {
	value, err := rand.Int(rand.Reader, big.NewInt(int64(maxInt)))
	if err != nil {
		mlog.Error("Error generating random number", mlog.Err(err))
		panic(err.Error())
	}
	return int(value.Int64())
}

func randomString(runes []rune, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[randomInt(len(runes))]
	}
	return string(b)
}

func generateUUIDName() string {
	id := uuid.New()
	return (id.String())
}

func generateTeamChannelName(teamName string, channelName string) string {
	name := teamName
	if name != "" {
		name += "-"
	}
	name += channelName
	name += "-" + randomString(LETTERS, 10)

	return name
}

func generatePersonalMeetingName(username string) string {
	return username + "-" + randomString(LETTERS, 20)
}

func generateRandomName() string {
	return hotdiva2000.Generate()
}
