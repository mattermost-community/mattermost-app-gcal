package function

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

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
			fieldUseServiceAccount,
			fieldImpersonateEmail,
		},
	},

	Handler: RequireGoogleToken(func(creq CallRequest) apps.CallResponse {
		message := ""
		oauth2Service, err := oauth2api.NewService(creq.ctx, option.WithTokenSource(creq.ts))
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get OAuth2 service"))
		}
		uiService := oauth2api.NewUserinfoService(oauth2Service)
		ui, err := uiService.V2.Me.Get().Do()
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get user info"))
		}
		message += fmt.Sprintf("#### User info:\n%s\n", utils.JSONBlock(ui))

		calID := creq.GetValue(fCalendarID, "")
		calService, err := calendar.NewService(creq.ctx, WithCommandAuth(creq))
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
		}

		cal, err := calService.Calendars.Get(calID).Do()
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get calendar"))
		}
		message += fmt.Sprintf("#### Calendar:\n%s\n", utils.JSONBlock(cal))

		events, err := calService.Events.List(calID).Do()
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to list events"))
		}

		message += fmt.Sprintf("#### Events:\n%s\n", utils.JSONBlock(events))
		return apps.NewTextResponse(message)
	}),
}
