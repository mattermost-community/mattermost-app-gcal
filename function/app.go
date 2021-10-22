package function

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"go.uber.org/zap/zapcore"
	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const IconPath = "icon.png"

var AppPathPrefix = ""
var Log = utils.MustMakeCommandLogger(zapcore.DebugLevel)

type CallRequest struct {
	apps.CallRequest

	ts  oauth2.TokenSource
	ctx context.Context
	log utils.Logger
}

func Init() {
	// Ping
	http.HandleFunc(AppPathPrefix+"/ping",
		httputils.HandleJSONData([]byte("{}")))

	// Bindings
	HandleCall("/bindings", bindings)

	// OAuth2 (Google Calendar) connect commands and callbacks.
	HandleCall("/oauth2/connect", oauth2Connect)
	HandleCall("/oauth2/complete", oauth2Complete)

	// Google Calendar webhook handler
	HandleCall(appspath.Webhook, webhookReceived)

	// Commands
	HandleSimpleCommand(configure)
	HandleSimpleCommand(connect)
	HandleSimpleCommand(disconnect)
	HandleSimpleCommand(debug)
	HandleSimpleCommand(subscribe)

	// Modals
	HandleCall("/configure-modal/submit",
		RequireAdmin(handleConfigureModal))

	// Lookups
	HandleCall(subscribe.Path+"/lookup",
		RequireGoogleToken(LookupHandler(lookupSubscribeCalendar)))

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		Log.Warnw("not found", "path", req.URL.Path, "method", req.Method)
		http.Error(w, fmt.Sprintf("Not found: %s %q", req.Method, req.URL.Path), http.StatusNotFound)
	})

}

type HandlerFunc func(CallRequest) apps.CallResponse

func FormHandler(h func(CallRequest) apps.Form) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		f := h(creq)
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

func HandleSimpleCommand(command SimpleCommand) {
	HandleCall(command.Submit.Path+"/submit", command.Handler)
}

func RequireAdmin(h HandlerFunc) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		if creq.Context.AdminAccessToken == "" {
			return apps.NewErrorResponse(
				utils.NewUnauthorizedError("missing admin access token, required to invoke " + creq.Path))
		}
		return h(creq)
	}
}

func doHandleCall(w http.ResponseWriter, req *http.Request, h HandlerFunc) {
	creq := CallRequest{}
	creq.ctx = context.Background()
	creq.log = Log.With("path", creq.Path)
	err := json.NewDecoder(req.Body).Decode(&creq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
