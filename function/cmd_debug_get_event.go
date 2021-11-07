package function

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var debugGetEvent = Command{
	Name:        "get-event",
	Description: "Get a specific event from a Google Calendar",

	BaseSubmit: apps.Call{
		Expand: &apps.Expand{
			OAuth2User: apps.ExpandAll,
			OAuth2App:  apps.ExpandAll,
		},
	},

	BaseForm: apps.Form{
		Fields: []apps.Field{
			fieldCalendarID(true, 1),
			fieldEventID(true, 2),
		},
	},

	Handler: RequireGoogleAuth(func(creq CallRequest) apps.CallResponse {
		outJSON := creq.BoolValue(fJSON)
		calID := creq.GetValue(fCalendarID, "")
		eventID := creq.GetValue(fEventID, "")
		calService, err := calendar.NewService(creq.ctx, creq.authOption)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
		}

		e, err := calService.Events.Get(calID, eventID).Do()
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get event"))
		}

		message := fmt.Sprintf("#### %s\n", EventSummaryString(e))
		message += fmt.Sprintf("  Time: %s\n", EventDateTimeString(e))
		if len(e.Attendees) > 0 {
			message += fmt.Sprintf("  Guests: %s\n", EventAttendeesString(e))
		}
		if e.Description != "" {
			message += fmt.Sprintf("  %s\n", e.Description)
		}

		if outJSON {
			message += "----\n"
			message += utils.JSONBlock(e)
		}

		return apps.NewTextResponse(message)
	}),
}
