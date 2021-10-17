package function

import (
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func oauth2Config(creq CallRequest) *oauth2.Config {
	fmt.Printf("<>/<> %s: getOAuth2Config: %s", creq.Path, utils.Pretty(creq.Context.OAuth2))
	return &oauth2.Config{
		ClientID:     creq.Context.OAuth2.ClientID,
		ClientSecret: creq.Context.OAuth2.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  creq.Context.OAuth2.CompleteURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}
}

var connect = command{
	Meta: Meta{
		Name: "connect",
	},

	call: apps.Call{
		Expand: &apps.Expand{
			OAuth2App: apps.ExpandAll,
		},
	},

	handler: func(creq CallRequest) apps.CallResponse {
		fmt.Printf("<>/<> %s: handler: context: %s", creq.Path, utils.Pretty(creq.Context))
		return apps.NewTextResponse("[Connect](%s) to Google.", creq.Context.OAuth2.ConnectURL)
	},
}

var disconnect = command{
	Meta: Meta{
		Name: "disconnect",
	},

	call: apps.Call{
		Expand: &apps.Expand{
			ActingUserAccessToken: apps.ExpandAll,
		},
	},

	handler: func(creq CallRequest) apps.CallResponse {
		asActingUser := appclient.AsActingUser(creq.Context)
		err := asActingUser.StoreOAuth2User(creq.Context.AppID, nil)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to reset the stored user"))
		}
		return apps.NewTextResponse("Disconnected your Google account.")
	},
}

func oauth2Connect(creq CallRequest) apps.CallResponse {
	state, _ := creq.Values["state"].(string)
	url := oauth2Config(creq).AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return apps.NewDataResponse(url)
}

func oauth2Complete(creq CallRequest) apps.CallResponse {
	code, _ := creq.Values["code"].(string)
	token, err := oauth2Config(creq).Exchange(creq.ctx, code)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed token exchange"))
	}

	err = StoreOAuth2User(creq, token)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to store OAuth user info to Mattermost"))
	}
	return apps.NewTextResponse("completed connecting to Google Calendar with OAuth2.")
}

// func RefreshOAuth2Token(creq CallRequest) oauth2.TokenSource {
// 	oauthConfig := oauth2Config(creq)
// 	token := oauth2.Token{}
// 	remarshal(&token, creq.Context.OAuth2.User)
// 	tokenSource := oauthConfig.TokenSource(creq.ctx, &token)

// 	// Store new token if refreshed
// 	newToken, err := tokenSource.Token()
// 	if err != nil && newToken.AccessToken != token.AccessToken {
// 		_ = StoreOAuth2User(creq, newToken)
// 	}

// 	return tokenSource
// }

func StoreOAuth2User(creq CallRequest, token *oauth2.Token) error {
	asActingUser := appclient.AsActingUser(creq.Context)
	return asActingUser.StoreOAuth2User(creq.Context.AppID, token)
}
