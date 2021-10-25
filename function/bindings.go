package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"golang.org/x/oauth2"
)

func bindings(creq CallRequest) apps.CallResponse {

	bindings := apps.Binding{
		Icon:        IconPath,
		Label:       "gcal",
		Description: "Google Calendar Mattermost App",
	}

	token := oauth2.Token{}
	remarshal(&token, creq.Context.OAuth2.User)
	if token.AccessToken == "" {
		bindings.Bindings = append(bindings.Bindings,
			connect.Binding)
	} else {
		bindings.Bindings = append(bindings.Bindings,
			debug.Binding,
			subscribe.Binding,
			disconnect.Binding,
		)
	}

	if creq.Context.ActingUser.IsSystemAdmin() {
		bindings.Bindings = append(bindings.Bindings,
			configure.Binding)
		bindings.Bindings = append(bindings.Bindings,
			info.Binding)
	}

	return apps.NewDataResponse([]apps.Binding{
		{
			Location: apps.LocationCommand,
			Bindings: []apps.Binding{
				bindings,
			},
		},
	})
}
