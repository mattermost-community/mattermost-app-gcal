package function

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var debugListEvents = Command{
	Name:        "list-events",
	Description: "List events in a Google Calendar",

	BaseSubmit: apps.Call{
		Expand: &apps.Expand{
			OAuth2User: apps.ExpandAll,
			OAuth2App:  apps.ExpandAll,
		},
	},

	BaseForm: apps.Form{
		Title: "Test Google Cal",
		Fields: []apps.Field{
			fieldCalendarID(true, 1),
			fieldDebugUseServiceAccount,
			fieldDebugImpersonateEmail,
		},
	},

	Handler: RequireGoogleAuth(func(creq CallRequest) apps.CallResponse {
		outJSON := creq.BoolValue(fJSON)
		calID := creq.GetValue(fCalendarID, "")
		calService, err := calendar.NewService(creq.ctx, creq.authOption)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
		}

		events, err := calService.Events.List(calID).Do()
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to list events"))
		}
		if len(events.Items) == 0 {
			message := fmt.Sprintf("No Google Calendar Events found in %s.", events.Description)
			if outJSON {
				message += "\n----\n"
				message += utils.JSONBlock(events)
			}
			return apps.NewTextResponse(message)
		}

		message := "#### List of Google Calendar events."
		for _, item := range events.Items {
			message += fmt.Sprintf("- %s\n", EventSummaryString(item))
			if len(item.Attendees) > 0 {
				message += fmt.Sprintf("  Guests: %s\n", EventAttendeesString(item))
			}
			if item.Description != "" {
				message += fmt.Sprintf("  %s\n", item.Description)
			}
		}

		if outJSON {
			message += "----\n"
			message += utils.JSONBlock(events)
		}

		return apps.NewTextResponse(message)
	}),
}
