package function

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-server/v6/model"
)

func webhookReceived(creq CallRequest) apps.CallResponse {
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
	if err != nil && errors.Cause(err) != utils.ErrNotFound {
		return apps.NewErrorResponse(err)
	}

	// DM (personal) webhooks should impersonate the user to access the
	// calendar. Channel (shared) webhooks must use the unaltered service
	// account credentials.
	impersonate := ""
	if s.MattermostChannelID == "" {
		impersonate = s.GoogleEmail
	}
	opt, err := ServiceAccountFromRequest(creq).AuthOption(creq.ctx, impersonate)
	if err != nil {
		return apps.NewErrorResponse(
			errors.Wrap(err, "failed to authenticate to Google"))
	}

	calService, err := calendar.NewService(creq.ctx, opt)
	if err != nil {
		return apps.NewErrorResponse(
			errors.Wrap(err, "failed to get Calendar client to Google"))
	}
	if s.SubID != subID {
		resourceID, _ := headers["X-Goog-Resource-Id"].(string)
		creq.CleanupOrphantSub(subID, resourceID, s)
	}

	syncToken := s.SyncToken
	resourceState, _ := headers["X-Goog-Channel-Id"].(string)
	if resourceState == "sync" {
		syncToken = ""
	}
	events, err := calService.Events.List(s.CalendarID).SyncToken(syncToken).Do()
	if err != nil {
		return apps.NewErrorResponse(
			errors.Wrap(err, "failed to list events"))
	}

	postf := func(p model.Post) {
		asBot.DMPost(s.MattermostUserID, &p)
	}
	if s.MattermostChannelID != "" {
		postf = func(p model.Post) {
			p.ChannelId = s.MattermostChannelID
			asBot.CreatePost(&p)
		}
	}

	// Post the results.
	postf(model.Post{
		Message: fmt.Sprintf("received notification: %s", utils.JSONBlock(headers)),
	})
	postf(model.Post{
		Message: fmt.Sprintf("events: %s", utils.JSONBlock(events)),
	})

	// Update the sync token in the subscription.
	s.SyncToken = events.NextSyncToken
	err = creq.StoreSub(s)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to update sync token in subscription"))
	}

	return apps.NewTextResponse("OK")
}
