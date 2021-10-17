package function

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const IconPath = "icon.png"

type CallRequest struct {
	apps.CallRequest

	ts  oauth2.TokenSource
	ctx context.Context
}

func InitHTTP(prefix string) {
	// Ping
	http.HandleFunc(prefix+"/ping",
		httputils.HandleData("text/plain", []byte("{}")))

	// Bindings
	HandleCall(prefix, "/bindings", handler(bindings))

	// OAuth2 (Google Calendar) connect commands and callbacks.
	HandleCall(prefix, "/oauth2/connect", handler(oauth2Connect))
	HandleCall(prefix, "/oauth2/complete", handler(oauth2Complete))

	// Commands
	HandleCommands(prefix, allCommands...)
}

type Handler interface {
	Handle(CallRequest) apps.CallResponse
}

type handler func(CallRequest) apps.CallResponse

func (h handler) Handle(creq CallRequest) apps.CallResponse {
	return h(creq)
}

func HandleCall(prefix, p string, h Handler) {
	http.HandleFunc(prefix+p, func(w http.ResponseWriter, req *http.Request) {
		doHandleCall(w, req, h, nil)
	})
}

func HandleCommands(prefix string, commands ...CommandHandler) {
	for i := range commands {
		command := commands[i]
		meta := command.Metadata()
		http.HandleFunc(prefix+meta.Path+"/form",
			func(w http.ResponseWriter, req *http.Request) {
				h := func(creq CallRequest) apps.CallResponse {
					form := command.Form(creq)
					return apps.NewFormResponse(form)
				}
				doHandleCall(w, req, handler(h), &meta)
			})

		http.HandleFunc(prefix+meta.Path+"/submit",
			func(w http.ResponseWriter, req *http.Request) {
				doHandleCall(w, req, command, &meta)
			})
	}
}

func doHandleCall(w http.ResponseWriter, req *http.Request, h Handler, meta *Meta) {
	creq := CallRequest{}
	creq.ctx = context.Background()
	err := json.NewDecoder(req.Body).Decode(&creq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if meta != nil {
		if meta.RequireAdmin && creq.Context.AdminAccessToken == "" {
			http.Error(w, "missing admin access token, required to invoke "+creq.Path, http.StatusUnauthorized)
		}

		if meta.RequireGoogleToken {
			if creq.Context.OAuth2.User == "" {
				http.Error(w, "missing google token, required to invoke "+creq.Path, http.StatusUnauthorized)
			}
			oauthConfig := oauth2Config(creq)
			token := oauth2.Token{}
			remarshal(&token, creq.Context.OAuth2.User)
			fmt.Printf("<>/<> %s: doHandleCall: creq.OAuth2: %s", creq.Path, utils.Pretty(creq.Context.OAuth2))
			fmt.Printf("<>/<> %s: doHandleCall: token: %s", creq.Path, utils.Pretty(token))

			creq.ts = oauthConfig.TokenSource(creq.ctx, &token)

			// // Store new token if refreshed
			// newToken, err := tokenSource.Token()
			// if err != nil && newToken.AccessToken != token.AccessToken {
			// 	_ = appclient.AsActingUser(creq.Context).StoreOAuth2User(creq.Context.AppID, newToken)
			// }
		}
	}

	_ = httputils.WriteJSON(w, h.Handle(creq))

	// if creq.Context.OAuth2.User != nil {
	// 	RefreshOAuth2Token(creq)
	// }
}

func remarshal(dst, src interface{}) {
	data, _ := json.Marshal(src)
	json.Unmarshal(data, dst)
}
