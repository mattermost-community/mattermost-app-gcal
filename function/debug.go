package function

import (
	"fmt"

	"google.golang.org/api/calendar/v3"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/pkg/errors"
)

var debug = command{
	Meta: Meta{
		Name:               "debug",
		RequireGoogleToken: true,
		Description:        "Test Google Cal",
	},

	form: apps.Form{
		Title: "Test Google Cal",
	},

	handler: func(creq CallRequest) apps.CallResponse {
		oauth2Service, err := oauth2api.NewService(creq.ctx, option.WithTokenSource(creq.ts))
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get OAuth2 connection to Google"))
		}

		uiService := oauth2api.NewUserinfoService(oauth2Service)
		ui, err := uiService.V2.Me.Get().Do()
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get UI client to Google"))
		}
		message := fmt.Sprintf("Hello from Google, [%s](mailto:%s)!", ui.Name, ui.Email)

		calService, err := calendar.NewService(creq.ctx, option.WithTokenSource(creq.ts))
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
		}

		cl, err := calService.CalendarList.List().Do()
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get list of Calendars"))
		}
		if cl != nil && len(cl.Items) > 0 {
			message += " You have the following calendars:\n"
			for _, item := range cl.Items {
				message += "- " + item.Summary + "\n"
			}
		} else {
			message += " You have no calendars.\n"
		}

		return apps.NewTextResponse(message)
	},
}
