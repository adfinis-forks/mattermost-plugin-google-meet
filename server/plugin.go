package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	mmi18n "github.com/mattermost/mattermost/server/public/pluginapi/i18n"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
)

const gmeetNameSchemeAsk = "ask"
const gmeetNameSchemeWords = "words"
const gmeetNameSchemeUUID = "uuid"
const gmeetNameSchemeMattermost = "mattermost"

const configChangeEvent = "custom_gmeet_config_update"

const gmeetURL = "https://g.co/meet"

type UserConfig struct {
	NamingScheme string `json:"naming_scheme"`
}

type Plugin struct {
	plugin.MattermostPlugin

	client *pluginapi.Client

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	b *mmi18n.Bundle

	botID string
}

func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		return err
	}

	command, err := p.createGmeetCommand()
	if err != nil {
		return err
	}

	if err = p.API.RegisterCommand(command); err != nil {
		return err
	}

	i18nBundle, err := mmi18n.InitBundle(p.API, filepath.Join("assets", "i18n"))
	if err != nil {
		return err
	}
	p.b = i18nBundle

	gmeetBot := &model.Bot{
		Username:    "gmeet",
		DisplayName: "Google Meet",
		Description: "A bot account created by the google meet plugin",
	}
	options := []pluginapi.EnsureBotOption{
		pluginapi.ProfileImagePath("assets/icon.png"),
	}

	p.client = pluginapi.NewClient(p.API, p.Driver)
	botID, ensureBotError := p.client.Bot.EnsureBot(gmeetBot, options...)
	if ensureBotError != nil {
		return errors.Wrap(ensureBotError, "failed to ensure gmeet bot user")
	}

	p.botID = botID

	return nil
}

func (p *Plugin) startMeeting(user *model.User, channel *model.Channel, meetingID string, meetingTopic string, _ bool, rootID string) (string, error) {
	l := p.b.GetServerLocalizer()
	if meetingID == "" {
		meetingID = encodeGmeetMeetingID(meetingTopic)
		if meetingID != "" {
			meetingID += "-"
		}
		meetingID += randomString(LETTERS, 20)
	}
	meetingPersonal := false
	defaultMeetingTopic := p.b.LocalizeDefaultMessage(l, &i18n.Message{
		ID:    "gmeet.start_meeting.default_meeting_topic",
		Other: "Google Meeting",
	})

	if len(meetingTopic) < 1 {
		userConfig, err := p.getUserConfig(user.Id)
		if err != nil {
			return "", err
		}

		switch userConfig.NamingScheme {
		case gmeetNameSchemeWords:
			meetingID = generateRandomName()
		case gmeetNameSchemeUUID:
			meetingID = generateUUIDName()
		case gmeetNameSchemeMattermost:
			if channel.Type == model.ChannelTypeDirect || channel.Type == model.ChannelTypeGroup {
				meetingID = generatePersonalMeetingName(user.Username)
				meetingTopic = p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "gmeet.start_meeting.personal_meeting_topic",
						Other: "{{.Name}}'s Personal Meeting",
					},
					TemplateData: map[string]string{"Name": user.GetDisplayName(model.ShowNicknameFullName)},
				})
				meetingPersonal = true
			} else {
				team, teamErr := p.API.GetTeam(channel.TeamId)
				if teamErr != nil {
					return "", teamErr
				}
				meetingTopic = p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "gmeet.start_meeting.channel_meeting_topic",
						Other: "{{.ChannelName}} Channel Meeting",
					},
					TemplateData: map[string]string{"ChannelName": channel.DisplayName},
				})
				meetingID = generateTeamChannelName(team.Name, channel.Name)
			}
		default:
			meetingID = generateRandomName()
		}
	}

	meetingURL := gmeetURL + "/" + meetingID
	meetingLink := meetingURL

	meetingTypeString := p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "gmeet.start_meeting.meeting_id",
			Other: "Meeting ID",
		},
	})
	if meetingPersonal {
		meetingTypeString = p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "gmeet.start_meeting.personal_meeting_id",
				Other: "Personal Meeting ID (PMI)",
			},
		})
	}

	slackMeetingTopic := meetingTopic
	if slackMeetingTopic == "" {
		slackMeetingTopic = defaultMeetingTopic
	}

	slackAttachment := model.SlackAttachment{
		Fallback: p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID: "gmeet.start_meeting.fallback_text",
				Other: `Video Meeting started at [{{.MeetingID}}]({{.MeetingURL}}).

[Join Meeting]({{.MeetingURL}})`,
			},
			TemplateData: map[string]string{
				"MeetingID":  meetingID,
				"MeetingURL": meetingURL,
			},
		}),
		Title: slackMeetingTopic,
		Text: p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID: "gmeet.start_meeting.slack_attachment_text",
				Other: `{{.MeetingType}}: [{{.MeetingID}}]({{.MeetingURL}})

[Join Meeting]({{.MeetingURL}})`,
			},
			TemplateData: map[string]string{
				"MeetingType": meetingTypeString,
				"MeetingID":   meetingID,
				"MeetingURL":  meetingURL,
			},
		}),
	}

	if meetingTopic == "" {
		meetingTopic = meetingID
	}

	post := &model.Post{
		UserId:    user.Id,
		ChannelId: channel.Id,
		Type:      "custom_gmeet_post_type",
		Props: map[string]interface{}{
			"attachments":           []*model.SlackAttachment{&slackAttachment},
			"meeting_id":            meetingID,
			"meeting_link":          meetingLink,
			"meeting_personal":      meetingPersonal,
			"meeting_topic":         meetingTopic,
			"default_meeting_topic": defaultMeetingTopic,
		},
		RootId: rootID,
	}

	if _, err := p.API.CreatePost(post); err != nil {
		return "", err
	}

	return meetingID, nil
}

func encodeGmeetMeetingID(meeting string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9-_]+")
	meeting = strings.ReplaceAll(meeting, " ", "-")
	return reg.ReplaceAllString(meeting, "")
}

func (p *Plugin) askMeetingType(user *model.User, channel *model.Channel, rootID string) error {
	l := p.b.GetUserLocalizer(user.Id)
	apiURL := *p.API.GetConfig().ServiceSettings.SiteURL + "/plugins/gmeet/api/v1/meetings"

	actions := []*model.PostAction{}

	var team *model.Team
	if channel.TeamId != "" {
		team, _ = p.API.GetTeam(channel.TeamId)
	}

	randomName := generateRandomName()
	actions = append(actions, &model.PostAction{
		Name: p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "gmeet.ask.meeting_name_random_words",
				Other: "Meeting name with random words",
			},
		}),
		Integration: &model.PostActionIntegration{
			URL: apiURL,
			Context: map[string]interface{}{
				"meeting_id":    randomName,
				"meeting_topic": randomName,
				"personal":      true,
			},
		},
	})

	actions = append(actions, &model.PostAction{
		Name: p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "gmeet.ask.personal_meeting",
				Other: "Personal meeting",
			},
		}),
		Integration: &model.PostActionIntegration{
			URL: apiURL,
			Context: map[string]interface{}{
				"meeting_id":    generatePersonalMeetingName(user.Username),
				"meeting_topic": fmt.Sprintf("%s's Meeting", user.GetDisplayName(model.ShowNicknameFullName)),
				"personal":      true,
			},
		},
	})

	if channel.Type == model.ChannelTypeOpen || channel.Type == model.ChannelTypePrivate {
		actions = append(actions, &model.PostAction{
			Name: p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "gmeet.ask.channel_meeting",
					Other: "Channel meeting",
				},
			}),
			Integration: &model.PostActionIntegration{
				URL: apiURL,
				Context: map[string]interface{}{
					"meeting_id":    generateTeamChannelName(team.Name, channel.Name),
					"meeting_topic": fmt.Sprintf("%s Channel Meeting", channel.DisplayName),
					"personal":      false,
				},
			},
		})
	}

	actions = append(actions, &model.PostAction{
		Name: p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "gmeet.ask.uuid_meeting",
				Other: "Meeting name with UUID",
			},
		}),
		Integration: &model.PostActionIntegration{
			URL: apiURL,
			Context: map[string]interface{}{
				"meeting_id":    generateUUIDName(),
				"meeting_topic": "Google Meeting",
				"personal":      false,
			},
		},
	})

	sa := model.SlackAttachment{
		Title: p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "gmeet.ask.title",
				Other: "Google Meeting Start",
			},
		}),
		Text: p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "gmeet.ask.select_meeting_type",
				Other: "Select type of meeting you want to start",
			},
		}),
		Actions: actions,
	}

	post := &model.Post{
		UserId:    p.botID,
		ChannelId: channel.Id,
		RootId:    rootID,
	}
	post.SetProps(map[string]interface{}{
		"attachments": []*model.SlackAttachment{&sa},
	})
	_ = p.API.SendEphemeralPost(user.Id, post)

	return nil
}

func (p *Plugin) deleteEphemeralPost(userID, postID string) {
	p.API.DeleteEphemeralPost(userID, postID)
}

func (p *Plugin) getUserConfig(userID string) (*UserConfig, error) {
	data, appErr := p.API.KVGet("config_" + userID)
	if appErr != nil {
		return nil, appErr
	}

	if data == nil {
		return &UserConfig{
			NamingScheme: p.getConfiguration().GmeetNamingScheme,
		}, nil
	}

	var userConfig UserConfig
	err := json.Unmarshal(data, &userConfig)
	if err != nil {
		return nil, err
	}

	return &userConfig, nil
}

func (p *Plugin) setUserConfig(userID string, config *UserConfig) error {
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}

	appErr := p.API.KVSet("config_"+userID, b)
	if appErr != nil {
		return appErr
	}

	p.API.PublishWebSocketEvent(configChangeEvent, nil, &model.WebsocketBroadcast{UserId: userID})
	return nil
}
