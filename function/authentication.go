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

func oauth2Connect(creq CallRequest) apps.CallResponse {
	state := creq.GetValue(fState, "")
	url := oauth2Config(creq).AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return apps.NewDataResponse(url)
}

func oauth2Complete(creq CallRequest) apps.CallResponse {
	code := creq.GetValue("code", "")
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

type ServiceAccount struct {
	// Mode is either "api_key", "service_account", or "" implying no service
	// account is to be used, and the corresponding functionality disabled.
	Mode        string `json:"mode"`              // fMode
	APIKey      string `json:"api_key,omitempty"` // fAPIKey
	AccountJSON string `json:"account_json"`      // fAccountJSON
}

func NewServiceAccount(mode, apiKey, serviceAccout string) ServiceAccount {
	sa := ServiceAccount{
		Mode: mode,
	}
	switch mode {
	case fAPIKey:
		sa.APIKey = apiKey
	case fAccountJSON:
		sa.AccountJSON = serviceAccout
	}
	return sa
}

func ServiceAccountFromRequest(creq CallRequest) ServiceAccount {
	m, _ := creq.Context.OAuth2.OAuth2App.Data.(map[string]interface{})
	mode, _ := m[fMode].(string)
	apiKey, _ := m[fAPIKey].(string)
	serviceAccount, _ := m[fAccountJSON].(string)
	return NewServiceAccount(mode, apiKey, serviceAccount)
}

func (sa ServiceAccount) String() string {
	switch sa.Mode {
	case fAPIKey:
		return fmt.Sprintf("API Key, ending in %q", utils.LastN(sa.APIKey, 8))
	case fAccountJSON:
		return fmt.Sprintf("Account JSON, ending in %q", utils.LastN(sa.APIKey, 8))
	default:
		return "No service account"
	}
}
