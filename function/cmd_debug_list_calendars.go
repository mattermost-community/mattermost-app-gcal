package function

import (
	"fmt"

	"google.golang.org/api/calendar/v3"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
)

var debugListCalendars = Command{
	Name:        "list-calendars",
	Description: "List Google Calendars",

	BaseSubmit: apps.Call{
		Expand: &apps.Expand{
			OAuth2User: apps.ExpandAll,
			OAuth2App:  apps.ExpandAll,
		},
	},

	BaseForm: apps.Form{
		Title: "Test Google Cal",
		Fields: []apps.Field{
			fieldUseServiceAccount,
			fieldImpersonateEmail,
		},
	},

	Handler: RequireGoogleToken(func(creq CallRequest) apps.CallResponse {
		calService, err := calendar.NewService(creq.ctx, WithCommandAuth(creq))
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
		}

		cl, err := calService.CalendarList.List().Do()
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get list of Calendars"))
		}
		message := fmt.Sprintf("Calendar list:\n%s\n", utils.JSONBlock(cl))

		return apps.NewTextResponse(message)
	}),
}
