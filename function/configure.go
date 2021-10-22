package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
)

var configure = SimpleCommand{
	Name: "configure",
	Submit: apps.Call{
		Expand: &apps.Expand{
			AdminAccessToken: apps.ExpandAll,
		},
	},
	Form: apps.Form{
		Title: "Configure Google Calendar App",
	},

	Handler: RequireAdmin(FormHandler(func(creq CallRequest) apps.Form {
		// display the form interactively, to enter the service account JSON,
		// not possible in autocomplete, yet.
		return apps.Form{
			Call: &apps.Call{
				Path: "/configure-modal",
				Expand: &apps.Expand{
					AdminAccessToken: apps.ExpandAll,
				},
			},
			Title: "Configure Google Calendar App credentials",
			Header: "" +
				"- Client ID and Secret are needed for the Mattermost users to authenticate to Google. " +
				"  [documentation](https://TODO)" +
				"\n" +
				"- Service Account is needed for to process incoming webhooks from Google. " +
				"  [documentation](https://TODO)" +
				"\n\n",
			Icon: IconPath,
			Fields: []apps.Field{
				{
					Name:        "client_id",
					ModalLabel:  "Client ID",
					Type:        apps.FieldTypeText,
					Description: "Google Client ID for the Mattermost App.",
					IsRequired:  true,
				},
				{
					Type:        apps.FieldTypeText,
					TextSubtype: apps.TextFieldSubtypePassword,
					Name:        "client_secret",
					ModalLabel:  "Client Secret",
					Description: "Google Client Secret for the Mattermost App.",
					IsRequired:  true,
				},
				{
					Type:          apps.FieldTypeText,
					TextSubtype:   apps.TextFieldSubtypeTextarea,
					TextMaxLength: 32 * 1024,
					Name:          "service_account",
					ModalLabel:    "Service Account (JSON)",
					Description:   "Google Service Account for the Mattermost App. Please open the downloaded credentials JSON file and paste its contents here.",
					IsRequired:    true,
				},
			},
		}
	})),
}.Init()

func handleConfigureModal(creq CallRequest) apps.CallResponse {
	clientID, _ := creq.Values["client_id"].(string)
	clientSecret, _ := creq.Values["client_secret"].(string)
	serviceAccount, _ := creq.Values["service_account"].(string)

	asAdmin := appclient.AsAdmin(creq.Context)
	err := asAdmin.StoreOAuth2App(creq.Context.AppID, apps.OAuth2App{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Data:         serviceAccount,
	})
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to store Oauth2 configuration to Mattermost"))
	}

	return apps.NewTextResponse(
		"updated OAuth client credentials: %q, %q",
		utils.LastN(clientID, 8),
		utils.LastN(clientSecret, 4),
	)
}
