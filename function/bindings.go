package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"golang.org/x/oauth2"
)

var adminCommands = []CommandHandler{
	configure,
}
var connectedCommands = []CommandHandler{
	debug,
	disconnect,
}
var disconnectedCommands = []CommandHandler{
	connect,
}

var allCommands []CommandHandler

func init() {
	allCommands = append(allCommands, adminCommands...)
	allCommands = append(allCommands, connectedCommands...)
	allCommands = append(allCommands, disconnectedCommands...)
}

func bindings(creq CallRequest) apps.CallResponse {
	bindings := apps.Binding{
		Icon:        IconPath,
		Label:       "gcal",
		Description: "Google Calendar Mattermost App",
	}

	fromCommands := func(commands []CommandHandler) []apps.Binding {
		bindings := []apps.Binding{}
		for _, c := range commands {
			bindings = append(bindings, c.Binding(creq))
		}
		return bindings
	}

	token := oauth2.Token{}
	remarshal(&token, creq.Context.OAuth2.User)
	if token.AccessToken == "" {
		bindings.Bindings = append(bindings.Bindings, fromCommands(disconnectedCommands)...)
	} else {
		bindings.Bindings = append(bindings.Bindings, fromCommands(connectedCommands)...)
	}

	if creq.Context.ActingUser.IsSystemAdmin() {
		bindings.Bindings = append(bindings.Bindings, fromCommands(adminCommands)...)
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
