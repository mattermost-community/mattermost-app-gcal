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

	// debug in dev mode, or sysadmin
	if creq.Context.DeveloperMode ||
		creq.Context.ActingUser != nil && creq.Context.ActingUser.IsSystemAdmin() {
		bindings.Bindings = append(bindings.Bindings, debugBinding(creq))
	}

	// user commands
	if creq.Context.OAuth2.User == nil {
		bindings.Bindings = append(bindings.Bindings,
			connect.Binding(creq))
	} else {
		bindings.Bindings = append(bindings.Bindings,
			watchBinding(creq),
			disconnect.Binding(creq),
		)
	}

	// admin commands
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

func debugBinding(creq CallRequest) apps.Binding {
	return apps.Binding{
		Label:    "debug",
		Location: "debug",
		Bindings: []apps.Binding{
			debugGetEvent.Binding(creq),
			debugListCalendars.Binding(creq),
			debugListEvents.Binding(creq),
			debugUserInfo.Binding(creq),
		},
		Icon: IconPath,
	}
}

func watchBinding(creq CallRequest) apps.Binding {
	return apps.Binding{
		Label:    "watch",
		Location: "watch",
		Bindings: []apps.Binding{
			watchStart.Binding(creq),
			watchList.Binding(creq),
			watchStop.Binding(creq),
		},
	}
}
