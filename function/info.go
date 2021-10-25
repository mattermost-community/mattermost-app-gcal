package function

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var info = SimpleCommand{
	Name:        "info",
	Description: "Google Calendar App information",

	Submit: apps.Call{
		Expand: &apps.Expand{
			AdminAccessToken: apps.ExpandAll,
			OAuth2App:        apps.ExpandAll,
		},
	},

	Form: apps.Form{
		Title: "Google Calendar App information",
	},

	Handler: RequireAdmin(func(creq CallRequest) apps.CallResponse {
		message := "Google Calendar App"
		if BuildDate != "" {
			message += fmt.Sprintf(": %q %q", BuildDate, BuildHashShort)
		}
		message += "\n"

		message += fmt.Sprintf("- OAuth2 complete URL: `%s`\n", creq.Context.OAuth2.CompleteURL)

		return apps.NewTextResponse(message)
	}),
}.Init()
