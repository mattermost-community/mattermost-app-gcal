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

	creq.log.Debugf("<>/<> WEBHOOK 4: context: %s", utils.Pretty(creq.Context))

	data, _ := creq.Context.OAuth2.Data.(string)
	if data == "" {
		return apps.NewErrorResponse(errors.New(
			"no service account data to process Google webhook request, use `/gcal configure` to update"))
	}
	// conf, err := google.JWTConfigFromJSON([]byte(data), OAuth2Scopes...)
	// if err != nil {
	// 	return apps.NewErrorResponse(err)
	// }
	// creq.log.Debugf("<>/<> WEBHOOK 5: got service account conf")

	calService, err := calendar.NewService(creq.ctx,
		option.WithCredentialsJSON([]byte(data)),
		option.WithScopes(OAuth2Scopes...),
	)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to get Calendar client to Google"))
	}
	creq.log.Debugf("<>/<> WEBHOOK 6: got client")

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
