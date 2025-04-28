package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

type StartMeetingRequest struct {
	ChannelID string `json:"channel_id"`
	Topic     string `json:"topic"`
	Personal  bool   `json:"personal"`
	MeetingID int    `json:"meeting_id"`
}

type StartMeetingFromAction struct {
	model.PostActionIntegrationRequest
	Context struct {
		MeetingID    string `json:"meeting_id"`
		MeetingTopic string `json:"meeting_topic"`
		RootID       string `json:"root_id"`
		Personal     bool   `json:"personal"`
	} `json:"context"`
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch path := r.URL.Path; path {
	case "/api/v1/meetings":
		p.handleStartMeeting(w, r)
	case "/api/v1/config":
		p.handleConfig(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) handleConfig(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")

	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	config, err := p.getUserConfig(userID)
	if err != nil {
		mlog.Error("Error getting user config", mlog.Err(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(config)
	if err != nil {
		mlog.Error("Error marshaling the Config to json", mlog.Err(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		mlog.Warn("Unable to write response body", mlog.String("handler", "handleConfig"), mlog.Err(err))
	}
}

func (p *Plugin) handleStartMeeting(w http.ResponseWriter, r *http.Request) {
	if err := p.getConfiguration().IsValid(); err != nil {
		mlog.Error("Invalid plugin configuration", mlog.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := r.Header.Get("Mattermost-User-Id")

	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		mlog.Debug("Unable to the user", mlog.Err(appErr))
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req StartMeetingRequest
	var action StartMeetingFromAction

	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		mlog.Debug("Unable to read request body", mlog.Err(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err1 := json.NewDecoder(bytes.NewReader(bodyData)).Decode(&req)
	err2 := json.NewDecoder(bytes.NewReader(bodyData)).Decode(&action)
	if err1 != nil && err2 != nil {
		mlog.Debug("Unable to decode the request content as start meeting request or start meeting action")
		http.Error(w, "Unable to decode your request", http.StatusBadRequest)
		return
	}

	channelID := req.ChannelID
	if channelID == "" {
		channelID = action.ChannelId
	}

	if _, err := p.API.GetChannelMember(channelID, userID); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	channel, appErr := p.API.GetChannel(channelID)
	if appErr != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	userConfig, err := p.getUserConfig(userID)
	if err != nil {
		mlog.Error("Error getting user config", mlog.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if userConfig.NamingScheme == gmeetNameSchemeAsk && action.PostId == "" {
		err = p.askMeetingType(user, channel, "")
		if err != nil {
			mlog.Error("Error asking the user for meeting name type", mlog.Err(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte("OK"))
		if err != nil {
			mlog.Warn("Unable to write response body", mlog.String("handler", "handleStartMeeting"), mlog.Err(err))
		}
		return
	}

	var meetingID string
	if userConfig.NamingScheme == gmeetNameSchemeAsk && action.PostId != "" {
		meetingID, err = p.startMeeting(user, channel, action.Context.MeetingID, action.Context.MeetingTopic, action.Context.Personal, "")
		if err != nil {
			mlog.Error("Error starting a new meeting from ask response", mlog.Err(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.deleteEphemeralPost(action.UserId, action.PostId)
	} else {
		meetingID, err = p.startMeeting(user, channel, "", req.Topic, req.Personal, action.Context.RootID)
		if err != nil {
			mlog.Error("Error starting a new meeting", mlog.Err(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	b, err := json.Marshal(map[string]string{"meeting_id": meetingID})
	if err != nil {
		mlog.Error("Error marshaling the MeetingID to json", mlog.Err(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		mlog.Warn("Unable to write response body", mlog.String("handler", "handleStartMeeting"), mlog.Err(err))
	}
}
