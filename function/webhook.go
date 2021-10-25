package function

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-server/v6/model"
)

func webhookReceived(creq CallRequest) apps.CallResponse {
	creq.log.Debugf("<>/<> WEBHOOK 1: values: %s", utils.Pretty(creq.Values))
	creq.log.Debugf("<>/<> WEBHOOK 2: type of headers: %T", creq.Values["headers"])
	headers, ok := creq.Values["headers"].(map[string]interface{})
	if !ok {
		return apps.NewErrorResponse(utils.NewInvalidError(
			"no header found in the Google webhook request"))
	}

	asBot := appclient.AsBot(creq.Context)

	subID, ok := headers["X-Goog-Channel-Id"].(string)
	if !ok || subID == "" {
		return apps.NewErrorResponse(utils.NewInvalidError(
			"header X-Goog-Channel-Id not found in the Google webhook request"))
	}
	s := Sub{}
	err := asBot.KVGet(subID, SubPrefix, &s)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	creq.log.Debugf("<>/<> WEBHOOK 3: loaded sub: %s", utils.Pretty(s))

	calID, ok := headers["X-Goog-Resource-Id"].(string)
	if !ok || calID == "" {
		return apps.NewErrorResponse(utils.NewInvalidError(
			"header X-Goog-Resource-Id not found in the Google webhook request"))
	}

	sa := ServiceAccountFromRequest(creq)
	creq.log.Debugf("<>/<> WEBHOOK 4: sa: %s", utils.Pretty(sa))

	var authOpt option.ClientOption
	switch sa.Mode {
	case fAccountJSON:
		authOpt = option.WithCredentialsJSON([]byte(sa.AccountJSON))
	case fAPIKey:
		authOpt = option.WithAPIKey(sa.APIKey)
	default:
		return apps.NewErrorResponse(errors.New(
			"no service account available to process Google webhook request, use `/gcal configure` to update"))
	}

	calService, err := calendar.NewService(creq.ctx, authOpt, option.WithScopes(OAuth2Scopes...))
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
	}
	creq.log.Debugf("<>/<> WEBHOOK 6: got an %q client", sa.Mode)

	events, err := calService.Events.List(calID).Do()
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to get calendar events"))
	}
	creq.log.Debugf("<>/<> WEBHOOK 7: got %v events", len(events.Items))

	asBot.DMPost(s.MattermostUserID, &model.Post{
		Message: fmt.Sprintf("received notification: %s", utils.JSONBlock(headers)),
	})

	asBot.DMPost(s.MattermostUserID, &model.Post{
		Message: fmt.Sprintf("events: %s", utils.JSONBlock(events)),
	})

	return apps.NewTextResponse("OK")
}
