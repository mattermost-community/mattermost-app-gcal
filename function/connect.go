package function

import (
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var OAuth2Scopes = []string{
	"https://www.googleapis.com/auth/calendar",
	"https://www.googleapis.com/auth/userinfo.profile",
	"https://www.googleapis.com/auth/userinfo.email",
}


func oauth2Config(creq CallRequest) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     creq.Context.OAuth2.ClientID,
		ClientSecret: creq.Context.OAuth2.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  creq.Context.OAuth2.CompleteURL,
		Scopes:       OAuth2Scopes,
	}
}

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

func oauth2Connect(creq CallRequest) apps.CallResponse {
	state, _ := creq.Values["state"].(string)
	url := oauth2Config(creq).AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return apps.NewDataResponse(url)
}

func oauth2Complete(creq CallRequest) apps.CallResponse {
	code, _ := creq.Values["code"].(string)
	oauth2Config := oauth2Config(creq)

	token, err := oauth2Config.Exchange(creq.ctx, code)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed token exchange"))
	}

	err = StoreOAuth2User(creq, token)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to store OAuth user info to Mattermost"))
	}
	return apps.NewTextResponse("completed connecting to Google Calendar with OAuth2.")
}

func StoreOAuth2User(creq CallRequest, token *oauth2.Token) error {
	asActingUser := appclient.AsActingUser(creq.Context)
	return asActingUser.StoreOAuth2User(creq.Context.AppID, token)
}

func RequireGoogleToken(h HandlerFunc) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		if creq.Context.OAuth2.User == "" {
			return apps.NewErrorResponse(
				utils.NewUnauthorizedError("missing google token, required to invoke " + creq.Path))
		}
		oauthConfig := oauth2Config(creq)
		token := oauth2.Token{}
		remarshal(&token, creq.Context.OAuth2.User)

		creq.ts = oauthConfig.TokenSource(creq.ctx, &token)

		// Store new token if refreshed
		newToken, err := creq.ts.Token()
		if err == nil && newToken.AccessToken != token.AccessToken {
			_ = appclient.AsActingUser(creq.Context).StoreOAuth2User(creq.Context.AppID, newToken)
		}

		return h(creq)
	}
}
