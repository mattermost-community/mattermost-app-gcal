package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type Sub struct {
	SubID       string `json:"id"`
	CreatorID   string `json:"creator_id",omitempty`
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

func (creq CallRequest) LoadSub(subID string) (*Sub, error) {
	asBot := appclient.AsBot(creq.Context)
	var s *Sub
	err := asBot.KVGet(subID, KVSubPrefix, &s)
	if err != nil {
		return s, errors.Wrap(err, "failed to load subscription from Mattermost")
	}
	if s == nil {
		return nil, utils.NewNotFoundError(subID)
	}
	return s, nil
}

// StoreSub stores the sub key, and updates the corresponding sub index key
// making sure it's included.
func (creq CallRequest) StoreSub(s Sub, updateIndex bool) error {
	asBot := appclient.AsBot(creq.Context)
	log := creq.log.With("sub_id", s.SubID, "creator_id", s.CreatorID, "cal_id", s.CalendarID)

	_, err := asBot.KVSet(s.SubID, KVSubPrefix, s)
	if err != nil {
		return errors.Wrap(err, "failed to save subscription to Mattermost")
	}
	log.Debugw("stored sub")
	if !updateIndex {
		return nil
	}

	subIDs := map[string]interface{}{}
	err = asBot.KVGet(s.CreatorID, KVSubIndexPrefix, subIDs)
	if err != nil {
		return errors.Wrap(err, "failed to load prior list of subscriptions")
	}
	if _, ok := subIDs[s.SubID]; ok {
		// nothing more to do
		return nil
	}

	subIDs[s.SubID] = struct{}{}
	_, err = asBot.KVSet(s.CreatorID, KVSubIndexPrefix, subIDs)
	if err != nil {
		return errors.Wrap(err, "failed to update list of subscriptions in Mattermost")
	}
	log.Debugw("updated sub index")

	return nil
}

func (creq CallRequest) DeleteSub(subID, creatorID string) error {
	asBot := appclient.AsBot(creq.Context)
	log := creq.log.With("sub_id", subID, "creator_id", creatorID)
	err := asBot.KVDelete(subID, KVSubPrefix)
	if err != nil {
		return errors.Wrap(err, "failed to delete subscription from Mattermost")
	}
	log.Debugw("stored sub")

	subIDs := map[string]interface{}{}
	err = asBot.KVGet(creatorID, KVSubIndexPrefix, &subIDs)
	if err != nil {
		return errors.Wrap(err, "failed to load prior list of subscriptions")
	}
	if _, ok := subIDs[subID]; !ok {
		// nothing more to do
		return nil
	}

	delete(subIDs, subID)
	_, err = asBot.KVSet(creatorID, KVSubIndexPrefix, subIDs)
	if err != nil {
		return errors.Wrap(err, "failed to update list of subscriptions in Mattermost")
	}
	log.Debugw("updated sub index")
	return nil
}

func (creq CallRequest) ListSubs(creatorID string) ([]Sub, error) {
	asBot := appclient.AsBot(creq.Context)
	subIDs := map[string]interface{}{}
	err := asBot.KVGet(creatorID, KVSubIndexPrefix, &subIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load list of subscriptions")
	}

	subs := []Sub{}
	for subID := range subIDs {
		sub, err := creq.LoadSub(subID)
		if err != nil {
			return nil, err
		}
		subs = append(subs, *sub)
	}
	return subs, nil
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
