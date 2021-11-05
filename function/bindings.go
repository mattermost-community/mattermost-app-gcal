package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func bindings(creq CallRequest) apps.CallResponse {
	bindings := apps.Binding{
		Icon:        IconPath,
		Label:       "gcal",
		Description: "Google Calendar Mattermost App",
	}

	if creq.Context.OAuth2.User == nil {
		bindings.Bindings = append(bindings.Bindings,
			connect.Binding(creq))
	} else {
		bindings.Bindings = append(bindings.Bindings,
			apps.Binding{
				Label: "debug",
				Bindings: []apps.Binding{
					debugListCalendars.Binding(creq),
					debugListEvents.Binding(creq),
				},
				Icon: IconPath,
			},
			subscribe.Binding(creq),
			disconnect.Binding(creq),
		)
	}

	if creq.Context.ActingUser != nil && creq.Context.ActingUser.IsSystemAdmin() {
		bindings.Bindings = append(bindings.Bindings,
			configure.Binding(creq))
		bindings.Bindings = append(bindings.Bindings,
			info.Binding(creq))
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
