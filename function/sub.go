package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type Sub struct {
	SubID       string `json:"id"`
	GoogleEmail string `json:"google_user_id"`
	CalendarID  string `json:"calendar_id,omitempty"`

	Google    *calendar.Channel `json:"google,omitempty"`
	SyncToken string            `json:"next_sync_token,omitempty"`

	// Subscription destination: MattermostUserID for a personal subscription
	// via DMs, or a ChannelID to post into a channel (as bot). If both are
	// specified, the subscription is trated as personal.
	MattermostUserID    string `json:"mattermost_user_id"`
	MattermostChannelID string `json:"channel_id,omitempty"`
}

func (creq CallRequest) StoreSub(s Sub) error {
	asBot := appclient.AsBot(creq.Context)
	_, err := asBot.KVSet(s.SubID, SubPrefix, s)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to save subscription to Mattermost"))
	}
	return nil
}

func (creq CallRequest) DeleteSub(id string) error {
	asBot := appclient.AsBot(creq.Context)
	err := asBot.KVDelete(id, SubPrefix)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to delete subscription from Mattermost"))
	}
	return nil
}

func (creq CallRequest) CleanupOrphantSub(subID, resourceID string, s Sub) {
	creq.log.Infof(`Subscription %s is not found or is not valid. Cleaning up fails with "googleapi: Error 403: Incorrect OAuth client, stopChannelClientIncorrect"`, subID)

	// err := creq.DeleteSub(subID)
	// if err != nil {
	// 	creq.log.WithError(err).Debugf("failed to clean up orphant subscription")
	// }

	// opt, err := ServiceAccountFromRequest(creq).AuthOption(creq.ctx, "lev@pacificolabs.com") // s.GoogleEmail)
	// if err != nil {
	// 	creq.log.WithError(err).Debugf("failed to authenticate to Google")
	// 	return
	// }
	// calService, err := calendar.NewService(creq.ctx, opt)
	// if err != nil {
	// 	creq.log.WithError(err).Debugf("failed to get Calendar client to Google")
	// 	return
	// }
	// err = calService.Channels.Stop(&calendar.Channel{
	// 	Id:         subID,
	// 	ResourceId: resourceID,
	// 	Address:    creq.appProxyURL(path.Webhook),
	// 	Type:       "web_hook",
	// 	Token:      "TODO",
	// }).Do()
	// if err != nil {
	// 	creq.log.WithError(err).Debugf("failed to stop orphant subscription")
	// 	return
	// }
}
