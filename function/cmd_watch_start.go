package function

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var watchStart = Command{
	Name:        "start",
	Description: "Start a personal subscription to Google Calendar change notifications",

	BaseSubmit: apps.Call{
		Expand: &apps.Expand{
			App:                   apps.ExpandSummary,
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
		},
	},

	BaseForm: apps.Form{
		Fields: []apps.Field{
			fieldCalendarID(true, 1),
		},
	},

	Handler: RequireGoogleAuth(
		func(creq CallRequest) apps.CallResponse {
			calID := creq.GetValue(fCalendarID, "")
			if calID == "" {
				return apps.NewErrorResponse(utils.NewInvalidError("no calendar ID provided"))
			}
			creq.log = creq.log.With("cal_id", calID)

			s := Sub{
				SubID:            model.NewId(),
				CalendarID:       calID,
				CreatorID:        creq.Context.ActingUserID,
				MattermostUserID: creq.Context.ActingUserID,
				GoogleEmail:      creq.user.Email,
			}

			// Pre-save the "incomplete" subscription record so that it is
			// available when we get the first sync webhook message.
			err := creq.StoreSub(&s, false)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to pre-save subscription"))
			}

			channelIn := &calendar.Channel{
				Id:      s.SubID,
				Address: creq.appProxyURL(path.Webhook),
				Type:    "web_hook",
				Token:   model.NewId(),
			}
			calService, err := calendar.NewService(creq.ctx, creq.authOption)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
			}
			channel, err := calService.Events.Watch(calID, channelIn).Do()
			if err != nil {
				return apps.NewErrorResponse(errors.Wrapf(err, "failed to watch %s"))
			}
			creq.log.Debugf("started watching:\n%s\nresponse: %s", utils.Pretty(channelIn), utils.Pretty(channel))

			// Save the subscription record again, this time complete with the Watch
			// response.
			s.Google = channel
			err = creq.StoreSub(&s, true)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to complete subscription"))
			}

			message := fmt.Sprintf("Subscibed to **%s**\n- ID:`%s`\n- Google resource (channel) ID:`%s`\n- Sync token:`%s`\n", s.CalendarSummary, s.SubID, s.Google.ResourceId, s.SyncToken)
			return RespondWithJSON(creq, message, s)
		}),
}

func canWatchToChannel(creq CallRequest) bool {
	cc := creq.Context
	switch {
	case cc.ActingUser != nil && cc.ActingUser.IsSystemAdmin():
		return true

	case cc.TeamMember != nil && model.IsInRole(cc.TeamMember.Roles, model.TeamAdminRoleId):
		return true

	case cc.ChannelMember != nil && model.IsInRole(cc.ChannelMember.Roles, model.ChannelAdminRoleId):
		return true
	}
	return false
}
