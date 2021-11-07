package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type Sub struct {
	// SubID is a unique ID generated at the creation of the Sub. It is also
	// used as the channel ID on the Google side.
	SubID string `json:"id"`

	// CreatorID is the Mattermost ID of the user who created the subscription.
	// It is used to namespace the cached events for the user as they are stored
	// by the webhook handler.
	CreatorID string `json:"creator_id",omitempty`

	// GoogleEmail is the email ID of the Google user, used for impersonation
	// with a service account.
	GoogleEmail string `json:"google_user_id"`

	// CalendarID of the calendar subscribed to.
	CalendarID string `json:"calendar_id,omitempty"`

	// CalendarSummary is the summary (title) of the calendar, also used as the
	// Sub name. Updated on change every time Events.List is invoked, so should
	// be in sync with the actual name of the calendar.
	CalendarSummary string `json:"calendar_summary,omitempty"`

	// Google channel for the webhook.
	Google *calendar.Channel `json:"google,omitempty"`

	// SyncToken to use for the the next update.
	SyncToken string `json:"next_sync_token,omitempty"`

	// MattermostUserID to DM on updates.
	// TODO: support subscribing to channels
	MattermostUserID string `json:"mattermost_user_id"`
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
func (creq CallRequest) StoreSub(s *Sub, updateIndex bool) error {
	asBot := appclient.AsBot(creq.Context)

	_, err := asBot.KVSet(s.SubID, KVSubPrefix, s)
	if err != nil {
		return errors.Wrap(err, "failed to save subscription to Mattermost")
	}
	if !updateIndex {
		return nil
	}

	var subIDs map[string]interface{}
	err = asBot.KVGet(s.CreatorID, KVSubIndexPrefix, &subIDs)
	if err != nil {
		return errors.Wrap(err, "failed to load prior list of subscriptions")
	}
	if subIDs == nil {
		subIDs = map[string]interface{}{}
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

	return nil
}

func (creq CallRequest) DeleteSub(s *Sub) error {
	asBot := appclient.AsBot(creq.Context)
	log := creq.log.With("sub_id", s.SubID, "creator_id", s.CreatorID)
	err := asBot.KVDelete(s.SubID, KVSubPrefix)
	if err != nil {
		return errors.Wrap(err, "failed to delete subscription from Mattermost")
	}
	log.Debugw("stored sub")

	subIDs := map[string]interface{}{}
	err = asBot.KVGet(s.CreatorID, KVSubIndexPrefix, &subIDs)
	if err != nil {
		return errors.Wrap(err, "failed to load prior list of subscriptions")
	}
	if _, ok := subIDs[s.SubID]; !ok {
		// nothing more to do
		return nil
	}

	delete(subIDs, s.SubID)
	_, err = asBot.KVSet(s.CreatorID, KVSubIndexPrefix, subIDs)
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
