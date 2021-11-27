package function

import (
	"fmt"

	"google.golang.org/api/calendar/v3"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/pkg/errors"
)

var debugListCalendars = Command{
	Name:        "list-calendars",
	Description: "List Google Calendars",

	BaseSubmit: &apps.Call{
		Expand: &apps.Expand{
			OAuth2User: apps.ExpandAll,
			OAuth2App:  apps.ExpandAll,
		},
	},

	Handler: RequireGoogleAuth(
		func(creq CallRequest) apps.CallResponse {
			calService, err := calendar.NewService(creq.ctx, creq.authOption)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
			}
			cl, err := calService.CalendarList.List().Do()
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to get list of Calendars"))
			}

			if len(cl.Items) == 0 {
				return RespondWithJSON(creq, "No Google Calendars for this account", cl)
			}

			message := "#### List of Google Calendars:\n\n"

			for _, item := range cl.Items {
				message += "- "
				if item.Deleted {
					message += "~~"
				}
				message += fmt.Sprintf("**%s** `%s` (%s)", item.Summary, item.Id, item.AccessRole)
				if item.Hidden {
					message += ", hidden"
				}
				if item.Selected {
					message += ", selected"
				}
				if item.Deleted {
					message += "~~"
				}
				message += ".\n"

				if item.Description != "" {
					message += fmt.Sprintf("  %s\n", item.Description)
				}
				message += "\n"
			}

			return RespondWithJSON(creq, message, cl)
		}),
}
