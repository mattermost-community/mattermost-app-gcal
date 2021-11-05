package function

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var info = Command{
	Name:        "info",
	Description: "Google Calendar App information",

	BaseSubmit: apps.Call{
		Expand: &apps.Expand{
			ActingUser: apps.ExpandSummary,
			OAuth2App:  apps.ExpandAll,
			OAuth2User: apps.ExpandAll,
		},
	},

	BaseForm: apps.Form{
		Title: "Google Calendar App information",
	},

	Handler: RequireAdmin(func(creq CallRequest) apps.CallResponse {
		message := "Google Calendar App"
		if BuildDate != "" {
			message += fmt.Sprintf(": %q %q", BuildDate, BuildHashShort)
		}
		message += "\n\n"
		message += fmt.Sprintf("☑ OAuth2 complete URL: `%s`\n", creq.Context.OAuth2.CompleteURL)

		if creq.Context.OAuth2.ClientID != "" {
			message += "☑ OAuth2 App: configured.\n"
			if creq.Context.OAuth2.Data != nil {
				message += "☑ Service account to process webhooks: configured.\n"
			} else {
				message += "☐ Service account to process webhooks: not configured. Please use `/gcal configure`.\n"
			}
		} else {
			message += "☐ OAuth2 App: not configured. Please use `/gcal configure`.\n"
		}

		if creq.Context.OAuth2.User != nil {
			user := User{}
			remarshal(&user, creq.Context.OAuth2.User)
			message += fmt.Sprintf("☑ Connected to Google as %s.\n", user.Email)
		} else {
			message += fmt.Sprintf("☐ Not connected to Google. Click [here](%s) to connect.\n", creq.Context.OAuth2.ConnectURL)
		}

		return apps.NewTextResponse(message)
	}),
}
