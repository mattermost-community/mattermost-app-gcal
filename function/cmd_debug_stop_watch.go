package function

import (
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var debugStopWatch = Command{
	Name:        "stop-watch",
	Description: "Stop Google watch with the internal IDs",

	BaseSubmit: apps.Call{
		Expand: &apps.Expand{
			OAuth2User: apps.ExpandAll,
			OAuth2App:  apps.ExpandAll,
		},
	},

	BaseForm: apps.Form{
		Fields: []apps.Field{
			{
				Name:        fID,
				Type:        apps.FieldTypeText,
				Description: "Google channel ID (aka subscription ID, also see Channel)",
				IsRequired:  true,
			},
			{
				Name:        fResourceID,
				Type:        apps.FieldTypeText,
				Description: "Google resource ID (see Channel)",
				IsRequired:  true,
			},
		},
	},

	Handler: RequireGoogleAuth(
		func(creq CallRequest) apps.CallResponse {
			id := creq.GetValue(fID, "")
			resourceID := creq.GetValue(fResourceID, "")

			calService, err := calendar.NewService(creq.ctx, creq.authOption)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
			}

			err = calService.Channels.Stop(&calendar.Channel{
				Id:         id,
				ResourceId: resourceID,
			}).Do()
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to stop Google watch"))
			}

			return RespondWithJSON(creq, "Successfully unsubscribed: "+id, nil)
		}),
}
