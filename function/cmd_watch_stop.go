package function

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var watchStop = Command{
	Name:        "stop",
	Description: "Stop a personal subscription to Google Calendar change notifications",

	BaseSubmit: apps.Call{
		Expand: &apps.Expand{
			OAuth2User: apps.ExpandAll,
			OAuth2App:  apps.ExpandAll,
		},
	},

	BaseForm: apps.Form{
		Fields: []apps.Field{
			fieldSubscriptionID(true, 1),
		},
	},

	Handler: RequireGoogleAuth(
		func(creq CallRequest) apps.CallResponse {
			subID := creq.GetValue(fSubscriptionID, "")
			if subID == "" {
				return apps.NewErrorResponse(utils.NewInvalidError("no subscription ID provided"))
			}
			creq.log = creq.log.With("sub_id", subID)

			sub, err := creq.LoadSub(subID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			calService, err := calendar.NewService(creq.ctx, creq.authOption)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
			}

			err = calService.Channels.Stop(sub.Google).Do()
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to stop Google watch"))
			}

			err = creq.DeleteSub(sub)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to delete subscription record from Mattermost"))
			}

			return RespondWithJSON(creq, fmt.Sprintf("Successfully unsubscribed: %s", sub.CalendarSummary), sub)
		}),
}
