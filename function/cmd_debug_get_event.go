package function

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var debugGetEvent = Command{
	Name:        "get-event",
	Description: "Get a specific event from a Google Calendar",

	BaseForm: &apps.Form{
		Title:"debug get event",
		Submit: &apps.Call{
			Expand: &apps.Expand{
				OAuth2User: apps.ExpandAll,
				OAuth2App:  apps.ExpandAll,
			},
		},
		Fields: []apps.Field{
			fieldCalendarID(true, 1),
			fieldEventID(true, 2),
		},
	},

	Handler: RequireGoogleAuth(func(creq CallRequest) apps.CallResponse {
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

		return RespondWithJSON(creq, message, e)
	}),
}
