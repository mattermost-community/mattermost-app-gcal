package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
)

var configure = command{
	Meta: Meta{
		Name:         "configure",
		RequireAdmin: true,
	},

	form: apps.Form{
		Title: "Configures Google OAuth2 App credentials",
		Fields: []apps.Field{
			{
				Type:        "text",
				Name:        "client_id",
				Description: "Google Client ID for the Mattermost App",
				IsRequired:  true,
			},
			{
				Type:        "text",
				Name:        "client_secret",
				Description: "Google Client Secret for the Mattermost App",
				IsRequired:  true,
			},
		},
	},

	handler: func(creq CallRequest) apps.CallResponse {
		clientID, _ := creq.Values["client_id"].(string)
		clientSecret, _ := creq.Values["client_secret"].(string)

		asAdmin := appclient.AsAdmin(creq.Context)
		err := asAdmin.StoreOAuth2App(creq.Context.AppID, clientID, clientSecret)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to store Oauth2 configuration to Mattermost"))
		}

		return apps.NewTextResponse(
			"updated OAuth client credentials: %q, %q",
			utils.LastN(clientID, 8),
			utils.LastN(clientSecret, 4),
		)
	},
}
