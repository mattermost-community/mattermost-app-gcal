package function

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var OAuth2Scopes = []string{
	calendar.CalendarScope,
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
	url := oauth2Config(creq).AuthCodeURL(state, oauth2.AccessTypeOffline)
	return apps.NewDataResponse(url)
}

func oauth2Complete(creq CallRequest) apps.CallResponse {
	code := creq.GetValue("code", "")
	oauth2Config := oauth2Config(creq)

	token, err := oauth2Config.Exchange(creq.ctx, code)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed token exchange"))
	}

	oauth2Service, err := oauth2api.NewService(creq.ctx,
		option.WithTokenSource(oauth2Config.TokenSource(creq.ctx, token)))
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to get OAuth2 service"))
	}
	uiService := oauth2api.NewUserinfoService(oauth2Service)
	ui, err := uiService.V2.Me.Get().Do()
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to get user info"))
	}

	asActingUser := appclient.AsActingUser(creq.Context)
	err = asActingUser.StoreOAuth2User(creq.Context.AppID, User{
		Token: token,
		Email: ui.Email,
		ID:    ui.Id,
	})
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to store OAuth user info to Mattermost"))
	}
	return apps.NewTextResponse("completed connecting to Google Calendar with OAuth2.")
}

type ServiceAccount struct {
	// Mode is either "api_key", "account_json", or "" implying no service
	// account is to be used, and the corresponding functionality disabled.
	Mode        string `json:"mode,omitempty"`         // fMode
	APIKey      string `json:"api_key,omitempty"`      // fAPIKey
	AccountJSON string `json:"account_json,omitempty"` // fAccountJSON
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

func (sa ServiceAccount) String() string {
	switch sa.Mode {
	case fAPIKey:
		return fmt.Sprintf("API Key, ending in %q", utils.LastN(sa.APIKey, 8))
	case fAccountJSON:
		return fmt.Sprintf("Account JSON, ending in %q", utils.LastN(sa.AccountJSON, 32))
	default:
		return "No service account"
	}
}

func (sa ServiceAccount) AuthOption(ctx context.Context, userEmail string) (option.ClientOption, error) {
	switch sa.Mode {
	case fAccountJSON:
		config, err := google.JWTConfigFromJSON(
			[]byte(sa.AccountJSON),
			calendar.CalendarScope)
		if err != nil {
			return nil, fmt.Errorf("JWTConfigFromJSON: %v", err)
		}
		config.Subject = userEmail

		return option.WithTokenSource(config.TokenSource(ctx)), nil

	case fAPIKey:
		return option.WithAPIKey(sa.APIKey), nil

	default:
		return nil, errors.New("service account authentication is not configured")
	}
}
