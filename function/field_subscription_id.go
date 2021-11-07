package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func fieldSubscriptionID(isRequired bool, autocompletePosition int) apps.Field {
	return apps.Field{
		Type:                 apps.FieldTypeDynamicSelect,
		Name:                 fSubscriptionID,
		Description:          "Choose a personal subscription to a Google Calendar",
		IsRequired:           isRequired,
		AutocompletePosition: autocompletePosition,
	}
}

func handleSubscriptionIDLookup(
	filter func(sub Sub) (include bool),
) HandlerFunc {
	h := func(creq CallRequest) []apps.SelectOption {
		opts := []apps.SelectOption{}

		owner := creq.Context.ActingUserID
		subs, err := creq.ListSubs(owner)
		if err != nil {
			creq.log.WithError(err).Warnf("failed to get list of subscriptions.")
			return nil
		}

		for _, sub := range subs {
			if filter != nil {
				if !filter(sub) {
					continue
				}
			}
			opts = append(opts, apps.SelectOption{
				Label: sub.CalendarSummary,
				Value: sub.SubID,
			})
		}
		return opts
	}

	return LookupHandler(h)
}
