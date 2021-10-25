package function

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
)

var connect = SimpleCommand{
	Name: "connect",
	Submit: apps.Call{
		Expand: &apps.Expand{
			OAuth2App: apps.ExpandAll,
		},
	},

	Handler: func(creq CallRequest) apps.CallResponse {
		return apps.NewTextResponse("[Connect](%s) to Google.", creq.Context.OAuth2.ConnectURL)
	},
}.Init()

var disconnect = SimpleCommand{
	Name: "disconnect",

	Submit: apps.Call{
		Expand: &apps.Expand{
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
		},
	},

	Handler: RequireGoogleToken(func(creq CallRequest) apps.CallResponse {
		asActingUser := appclient.AsActingUser(creq.Context)
		err := asActingUser.StoreOAuth2User(creq.Context.AppID, nil)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to reset the stored user"))
		}
		return apps.NewTextResponse("Disconnected your Google account.")
	}),
}.Init()
