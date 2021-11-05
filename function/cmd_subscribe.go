package function

import (
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var subscribeCall = apps.Call{
	Expand: &apps.Expand{
		App:                   apps.ExpandSummary,
		ActingUserAccessToken: apps.ExpandAll,
		OAuth2User:            apps.ExpandAll,
		OAuth2App:             apps.ExpandAll,
	},
}

var subscribe = Command{
	Name:        "me",
	Description: "Receive calendar change notifications as direct messages",
	BaseSubmit:  subscribeCall,

	BaseForm: apps.Form{
		Title: "Receive calendar change notifications as direct messages",
		Fields: []apps.Field{
			fieldCalendarID(true, 1),
			apps.Field{
				Name: fChannel,
				Type: apps.FieldTypeBool,
			},
		},
	},

	Handler: RequireGoogleToken(
		func(creq CallRequest) apps.CallResponse {
			calID := creq.GetValue(fCalendarID, "")
			if calID == "" {
				return apps.NewErrorResponse(utils.NewInvalidError("no calendar ID provided"))
			}

			// subID := model.NewId()
			// s := Sub{
			// 	SubID:      subID,
			// 	Creator:    creq.Context.ActingUserID,
			// 	CalendarID: calID,
			// }

			// authAsServiceAccount, err := ServiceAccountFromRequest(creq).AuthOption(creq.ctx, "")
			// if err != nil {
			// 	return apps.NewErrorResponse(errors.Wrap(err, "failed to get service account client"))
			// }
			// calService, err := calendar.NewService(creq.ctx, authAsServiceAccount)
			// if err != nil {
			// 	return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
			// }

			// calService.CalendarList.Insert(&calendar.CalendarListEntry{

			// })

			// asChannel := creq.BoolValue(fChannel)
			// if asChannel {
			// 	// Add the Bot user to the team and the channel.
			// 	teamID := creq.Context.TeamID
			// 	channelID := creq.Context.ChannelID
			// 	if teamID == "" || channelID == "" {
			// 		return apps.NewErrorResponse(utils.NewInvalidError(
			// 			"Team and Channel IDs are not found in the request"))
			// 	}

			// 	asUser := appclient.AsActingUser(creq.Context)
			// 	_, _, err := asUser.AddTeamMember(teamID, creq.Context.BotUserID)
			// 	if err != nil {
			// 		return apps.NewErrorResponse(errors.Wrapf(err,
			// 			"failed to add bot %q to the current team (%s)", creq.Context.App.BotUsername, teamID))
			// 	}
			// 	_, _, err = asUser.AddChannelMember(channelID, creq.Context.BotUserID)
			// 	if err != nil {
			// 		return apps.NewErrorResponse(errors.Wrapf(err,
			// 			"failed to add bot %q as member of the current channel (%s)", creq.Context.App.BotUsername, channelID))
			// 	}

			// 	s.MattermostChannelID = channelID
			// } else {
			// 	s.MattermostUserID = creq.Context.ActingUserID
			// }

			// // Save the "incomplete" subscription record so that it is available
			// // when we get the first sync webhook.
			// err := creq.StoreSub(s)
			// if err != nil {
			// 	return apps.NewErrorResponse(errors.Wrap(err, "failed to store incomplete subscription"))
			// }

			// channelIn := &calendar.Channel{
			// 	Id:      subID,
			// 	Address: creq.appProxyURL(path.Webhook),
			// 	Type:    "web_hook",
			// 	Token:   "TODO",
			// }
			// channel, err := calService.Events.Watch(calID, channelIn).Do()
			// if err != nil {
			// 	return apps.NewErrorResponse(errors.Wrapf(err, "failed to watch %s"))
			// }
			// creq.log.Debugf("started watching:\n%s\nresponse: %s", utils.Pretty(channelIn), utils.Pretty(channel))

			// // Save the subscription record again, this time complete with the Watch
			// // response.
			// s.Google = channel
			// err = creq.StoreSub(s)
			// if err != nil {
			// 	return apps.NewErrorResponse(errors.Wrap(err, "failed to complete subscription"))
			// }

			return apps.NewTextResponse("Successfully subscribed:%s", utils.JSONBlock("")) // channel))
		}),
}

var subscribeList = Command{
	Name:        "list",
	Description: "List existing subscriptions",
	BaseSubmit:  subscribeCall,

	BaseForm: apps.Form{
		Title: "List my subscriptions",
	},

	Handler: RequireGoogleToken(
		func(creq CallRequest) apps.CallResponse {

			calID := creq.GetValue(fCalendarID, "")
			if calID == "" {
				return apps.NewErrorResponse(utils.NewInvalidError("no calendar ID provided"))
			}

			subID := model.NewId()
			s := Sub{
				SubID:       subID,
				GoogleEmail: creq.user.Email,
				CalendarID:  calID,
			}

			asChannel := creq.BoolValue(fChannel)
			if asChannel {
				// Add the Bot user to the team and the channel.
				teamID := creq.Context.TeamID
				channelID := creq.Context.ChannelID
				if teamID == "" || channelID == "" {
					return apps.NewErrorResponse(utils.NewInvalidError(
						"Team and Channel IDs are not found in the request"))
				}

				asUser := appclient.AsActingUser(creq.Context)
				_, _, err := asUser.AddTeamMember(teamID, creq.Context.BotUserID)
				if err != nil {
					return apps.NewErrorResponse(errors.Wrapf(err,
						"failed to add bot %q to the current team (%s)", creq.Context.App.BotUsername, teamID))
				}
				_, _, err = asUser.AddChannelMember(channelID, creq.Context.BotUserID)
				if err != nil {
					return apps.NewErrorResponse(errors.Wrapf(err,
						"failed to add bot %q as member of the current channel (%s)", creq.Context.App.BotUsername, channelID))
				}

				s.MattermostChannelID = channelID
			} else {
				s.MattermostUserID = creq.Context.ActingUserID
			}

			// Save the "incomplete" subscription record so that it is available
			// when we get the first sync webhook.
			err := creq.StoreSub(s)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to store incomplete subscription"))
			}

			calService, err := calendar.NewService(creq.ctx, option.WithTokenSource(creq.ts))
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
			}

			channelIn := &calendar.Channel{
				Id:      subID,
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
			err = creq.StoreSub(s)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to complete subscription"))
			}

			return apps.NewTextResponse("Successfully subscribed:%s", utils.JSONBlock(channel))
		}),
}
