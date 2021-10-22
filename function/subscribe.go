package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const SubPrefix = "s"

type Sub struct {
	MattermostUserID string            `json:"user_id,omitempty"`
	CalendarID       string            `json:"calendar_id,omitempty"`
	Google           *calendar.Channel `json:"google,omitempty"`
}

var subscribe = SimpleCommand{
	Name:        "subscribe",
	Description: "Test Google Cal",

	Submit: apps.Call{
		Expand: &apps.Expand{
			App:                   apps.ExpandSummary,
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
		},
	},

	Form: apps.Form{
		Title: "Subscribe to notifications from Google Calendar",
		Fields: []apps.Field{
			{
				Type:                 apps.FieldTypeDynamicSelect,
				Name:                 "calendar_id",
				Description:          "specific Google Calendar",
				IsRequired:           true,
				AutocompletePosition: 1,
			},
		},
	},

	Handler: RequireGoogleToken(func(creq CallRequest) apps.CallResponse {
		calID := creq.GetValue("calendar_id", "")
		if calID == "" {
			return apps.NewErrorResponse(utils.NewInvalidError("no calendar ID provided"))
		}

		// Add the Bot user to the team and the channel.
		// teamID := creq.Context.TeamID
		// channelID := creq.Context.ChannelID

		// asUser := appclient.AsActingUser(creq.Context)
		// _, _, err := asUser.AddTeamMember(teamID, creq.Context.BotUserID)
		// if err != nil {
		// 	return apps.NewErrorResponse(errors.Wrapf(err,
		// 		"failed to add bot %q to the current team (%s)", creq.Context.App.BotUsername, teamID))
		// }
		// _, _, err = asUser.AddChannelMember(channelID, creq.Context.BotUserID)
		// if err != nil {
		// 	return apps.NewErrorResponse(errors.Wrapf(err,
		// 		"failed to add bot %q as member of the current channel (%s)", creq.Context.App.BotUsername, channelID))
		// }

		// Save the subscription record so it exists when we get the sync
		// message.
		googleChannelID := model.NewId()
		asBot := appclient.AsBot(creq.Context)
		s := Sub{
			CalendarID:       calID,
			MattermostUserID: creq.Context.ActingUserID,
		}
		_, err := asBot.KVSet(googleChannelID, SubPrefix, s)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to save subscription to Mattermost"))
		}

		calService, err := calendar.NewService(creq.ctx, option.WithTokenSource(creq.ts))
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
		}

		channelIn := &calendar.Channel{
			Id:      googleChannelID,
			Address: creq.appProxyURL(path.Webhook),
			Type:    "web_hook",
			Token:   "TODO",
		}
		channel, err := calService.Events.Watch(calID, channelIn).Do()
		if err != nil {
			return apps.NewErrorResponse(errors.Wrapf(err, "failed to watch %s"))
		}
		creq.log.Debugf("started watching:\n%s\nresponse: %s", utils.Pretty(channelIn), utils.Pretty(channel))

		// Save the subscription record again, this time complete with the Watch
		// response.
		s.Google = channel
		_, err = asBot.KVSet(googleChannelID, SubPrefix, s)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to save completed subscription to Mattermost"))
		}

		return apps.NewTextResponse("Successfully subscribed:%s", utils.JSONBlock(channel))
	}),
}.Init()

func lookupSubscribeCalendar(creq CallRequest) []apps.SelectOption {
	opts := []apps.SelectOption{}
	calService, err := calendar.NewService(creq.ctx, option.WithTokenSource(creq.ts))
	if err != nil {
		creq.log.WithError(err).Warnf("failed to get Calendar client.")
		return nil
	}

	cl, err := calService.CalendarList.List().Do()
	if err != nil {
		creq.log.WithError(err).Warnf("failed to get the list of Google calendars.")
		return nil
	}

	for _, item := range cl.Items {
		if item.Deleted || item.Hidden {
			continue
		}
		opts = append(opts, apps.SelectOption{
			Label: item.Summary,
			Value: item.Id,
		})
	}
	return opts
}
