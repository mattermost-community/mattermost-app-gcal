package function

import (
	"fmt"
	"net/http"

	"go.uber.org/zap/zapcore"

	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const IconPath = "icon.png"

var AppPathPrefix = ""
var Log = utils.MustMakeCommandLogger(zapcore.DebugLevel)

var BuildHash string
var BuildHashShort string
var BuildDate string

// KV store: subscriptions
const (
	// Individual subscriptions are stored in the "s" namespace, as "[s]{id}",
	// where id is the generated ID of the subscription, also known as channel
	// ID in Google terms.
	KVSubPrefix = "s"

	// Indices of subscriptions are stored in the "si" namespace, as
	// "[si]{userID}". The global (channel) subscriptions are stored under the
	// "bot_subs".
	KVSubIndexPrefix = "si"

	// Individual events are stored in the "e" namespace, as
	// "[e]base64({googleEmail}/{calID}/{eventID})".
	KVEventPrefix = "e"

	// The name of the key that stores the list of global subscriptions.
	KVBotSubscriptionsKey = "bot_subs"
)

// Field names
const (
	fAccountJSON       = "account_json"
	fAPIKey            = "api_key"
	fCalendarID        = "calendar_id"
	fChannel           = "channel"
	fClientID          = "client_id"
	fClientSecret      = "client_secret"
	fEventID           = "event_id"
	fID                = "id"
	fImpersonateEmail  = "impersonate_email"
	fJSON              = "json"
	fMode              = "mode"
	fResourceID        = "resource_id"
	fState             = "state"
	fSubscriptionID    = "sub_id"
	fUseServiceAccount = "use_service_account"
)

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
	HandleCommand(configure)
	HandleCommand(connect)
	HandleCommand(debugGetEvent)
	HandleCommand(debugListCalendars)
	HandleCommand(debugListEvents)
	HandleCommand(debugStopWatch)
	HandleCommand(debugUserInfo)
	HandleCommand(disconnect)
	HandleCommand(info)
	HandleCommand(watchList)
	HandleCommand(watchStart)
	HandleCommand(watchStop)

	// Modals
	HandleCall("/configure-modal/submit",
		RequireAdmin(handleConfigureModal))
	HandleCall("/configure-modal/form",
		RequireAdmin(FormHandler(handleConfigureModalForm)))

	// Lookups TODO rework when the paths are decoupled from forms
	HandleCall(watchStart.Path()+"/lookup",
		RequireGoogleAuth(handleCalendarIDLookup(nil)))
	HandleCall(watchStop.Path()+"/lookup",
		RequireGoogleAuth(handleSubscriptionIDLookup(nil)))
	HandleCall(debugListEvents.Path()+"/lookup",
		RequireGoogleAuth(handleCalendarIDLookup(nil)))
	HandleCall(debugGetEvent.Path()+"/lookup",
		RequireGoogleAuth(handleGetEventLookup))

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		Log.Warnw("not found", "path", req.URL.Path, "method", req.Method)
		http.Error(w, fmt.Sprintf("Not found: %s %q", req.Method, req.URL.Path), http.StatusNotFound)
	})
}
