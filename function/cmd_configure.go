package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var configure = Command{
	Name: "configure",
	BaseSubmit: &apps.Call{
		Expand: &apps.Expand{
			ActingUser:            apps.ExpandSummary,
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
		},
	},

	// Handler will display the configure form as a modal, to enter the service
	// account JSON, not possible in autocomplete, yet.
	Handler: RequireAdmin(
		FormHandler(configureModalForm)),
}
