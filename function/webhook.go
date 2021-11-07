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
	//TODO: process webhooks async
	err := doWebhookReceived(creq)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	return apps.NewTextResponse("OK")
}

func doWebhookReceived(creq CallRequest) error {
	headers, ok := creq.Values["headers"].(map[string]interface{})
	if !ok {
		return utils.NewInvalidError("no header found in the Google webhook request")
	}

	subID, ok := headers["X-Goog-Channel-Id"].(string)
	if !ok || subID == "" {
		return utils.NewInvalidError("header X-Goog-Channel-Id not found in the Google webhook request")
	}
	creq.log = creq.log.With("sub_id", subID)
	s, err := creq.LoadSub(subID)
	if err != nil && errors.Cause(err) != utils.ErrNotFound {
		return err
	}
	creq.log = creq.log.With("google_email", s.GoogleEmail, "cal_id", s.CalendarID)

	opt, err := ServiceAccountFromRequest(creq).AuthOption(creq.ctx, s.GoogleEmail)
	if err != nil {
		return errors.Wrap(err, "failed to authenticate to Google")
	}
	calService, err := calendar.NewService(creq.ctx, opt)
	if err != nil {
		return errors.Wrap(err, "failed to get Calendar client to Google")
	}

	syncToken := s.SyncToken
	resourceState, _ := headers["X-Goog-Resource-State"].(string)
	if resourceState == "sync" {
		syncToken = ""
	}
	creq.log = creq.log.With("sync_token", s.SyncToken)

	//TODO: implement NextPageToken support
	events, err := calService.Events.List(s.CalendarID).SyncToken(syncToken).Do()
	if err != nil {
		return errors.Wrap(err, "failed to list events")
	}

	asBot := appclient.AsBot(creq.Context)
	for _, e := range events.Items {
		problems := []error{}
		prev, err := creq.LoadEvent(s.GoogleEmail, s.CalendarID, e.Id)
		if err != nil && errors.Cause(err) != utils.ErrNotFound {
			problems = append(problems, err)
		}

		message := EventDiffString(prev, e)

		if len(problems) > 0 {
			message += "\n----\n"
			message += fmt.Sprintf("Errors: %v", problems)
		}

		rootPostID := ""
		if prev != nil {
			rootPostID = prev.RootPostID
		}
		post, _ := asBot.DMPost(s.MattermostUserID, &model.Post{
			Message: message,
			RootId:  rootPostID,
		})
		if rootPostID == "" && post != nil {
			rootPostID = post.Id
		}
		err = creq.StoreEvent(s.GoogleEmail, s.CalendarID, &Event{
			Event:      e,
			RootPostID: rootPostID,
		})
		if err != nil {
			problems = append(problems, err)
		}
	}

	// Update the sync token in the subscription.
	s.SyncToken = events.NextSyncToken
	err = creq.StoreSub(*s, false)
	if err != nil {
		return errors.Wrap(err, "failed to update sync token in subscription")
	}

	return nil
}
