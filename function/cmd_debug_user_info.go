package function

import (
	"fmt"

	"github.com/pkg/errors"
	oauth2api "google.golang.org/api/oauth2/v2"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var debugUserInfo = Command{
	Name:        "user-info",
	Description: "Obtains user info from Google Calendar",

	BaseSubmit: apps.Call{
		Expand: &apps.Expand{
			OAuth2User: apps.ExpandAll,
			OAuth2App:  apps.ExpandAll,
		},
	},

	Handler: RequireGoogleAuth(func(creq CallRequest) apps.CallResponse {
		oauth2Service, err := oauth2api.NewService(creq.ctx, creq.authOption)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get OAuth2 service"))
		}
		uiService := oauth2api.NewUserinfoService(oauth2Service)

		ui, err := uiService.V2.Me.Get().Do()
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to get user info"))
		}

		message := "#### Google Calendar user info\n\n"
		message += fmt.Sprintf("- Email: `%s`\n", ui.Email)
		message += fmt.Sprintf("- Name: `%s %s, %s`\n", ui.Gender, ui.FamilyName, ui.GivenName)

		return RespondWithJSON(creq, message, ui)
	}),
}
