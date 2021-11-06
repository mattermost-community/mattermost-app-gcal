package function

import (
	"context"
	"encoding/json"
	"net/http"
	"path"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

type CallRequest struct {
	apps.CallRequest

	user       User
	authOption option.ClientOption
	ctx        context.Context
	log        utils.Logger
}

type HandlerFunc func(CallRequest) apps.CallResponse

func FormHandler(h func(CallRequest) (apps.Form, error)) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		f, err := h(creq)
		if err != nil {
			return apps.NewErrorResponse(err)
		}
		return apps.NewFormResponse(f)
	}
}

func LookupHandler(h func(CallRequest) []apps.SelectOption) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		opts := h(creq)
		return apps.NewLookupResponse(opts)
	}
}

func HandleCall(p string, h HandlerFunc) {
	http.HandleFunc(AppPathPrefix+p, func(w http.ResponseWriter, req *http.Request) {
		doHandleCall(w, req, h)
	})
}

func HandleCommand(command Command) {
	HandleCall(command.Path()+"/submit", command.Handler)
}

func RequireAdmin(h HandlerFunc) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		if creq.Context.ActingUser != nil && !creq.Context.ActingUser.IsSystemAdmin() {
			return apps.NewErrorResponse(
				utils.NewUnauthorizedError("system administrator role is required to invoke " + creq.Path))
		}
		return h(creq)
	}
}

func RequireGoogleAuth(h HandlerFunc) HandlerFunc {
	return RequireGoogleUser(func(creq CallRequest) apps.CallResponse {
		if !creq.BoolValue(fUseServiceAccount) {
			token := creq.user.Token
			if token == nil {
				return apps.NewErrorResponse(
					utils.NewUnauthorizedError("missing Google OAuth2 token in the user record, required for " + creq.Path + ". Please use `/apps connect` to connect your Google account."))
			}

			oauthConfig := oauth2Config(creq)
			ts := oauthConfig.TokenSource(creq.ctx, token)

			// Store new token if refreshed
			newToken, err := ts.Token()
			if err == nil && newToken.AccessToken != token.AccessToken {
				creq.user.Token = newToken
				_ = appclient.AsActingUser(creq.Context).StoreOAuth2User(creq.Context.AppID, creq.user)
			}

			creq.authOption = option.WithTokenSource(ts)
		} else {
			email := creq.GetValue(fImpersonateEmail, "")
			opt, err := ServiceAccountFromRequest(creq).AuthOption(creq.ctx, email)
			if err != nil {
				return apps.NewErrorResponse(
					errors.Wrap(err, "failed to get service account impersonation auth option"))
			}
			creq.authOption = opt
		}
		return h(creq)
	})
}

func RequireGoogleUser(h HandlerFunc) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		if creq.Context.OAuth2.User == nil {
			return apps.NewErrorResponse(
				utils.NewUnauthorizedError("missing user record, required for " + creq.Path + ". Please use `/apps connect` to connect your Google account."))
		}
		user := User{}
		remarshal(&user, creq.Context.OAuth2.User)
		creq.user = user
		return h(creq)
	}
}

func ServiceAccountFromRequest(creq CallRequest) ServiceAccount {
	m, _ := creq.Context.OAuth2.OAuth2App.Data.(map[string]interface{})
	mode, _ := m[fMode].(string)
	apiKey, _ := m[fAPIKey].(string)
	serviceAccount, _ := m[fAccountJSON].(string)
	return NewServiceAccount(mode, apiKey, serviceAccount)
}

func doHandleCall(w http.ResponseWriter, req *http.Request, h HandlerFunc) {
	creq := CallRequest{}
	creq.ctx = context.Background()
	err := json.NewDecoder(req.Body).Decode(&creq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	creq.log = Log.With("path", creq.Path)

	cresp := h(creq)
	if cresp.Type == apps.CallResponseTypeError {
		creq.log.WithError(cresp).Debugw("Call failed.")
	}
	_ = httputils.WriteJSON(w, cresp)
}

func remarshal(dst, src interface{}) {
	data, _ := json.Marshal(src)
	json.Unmarshal(data, dst)
}

func (creq CallRequest) appProxyURL(paths ...string) string {
	p := path.Join(append([]string{creq.Context.AppPath}, paths...)...)
	return creq.Context.MattermostSiteURL + p
}
