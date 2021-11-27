package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"google.golang.org/api/calendar/v3"
)

func fieldEventID(isRequired bool, autocompletePosition int) apps.Field {
	return apps.Field{
		Type:                 apps.FieldTypeDynamicSelect,
		Name:                 fEventID,
		Description:          "Choose an Event",
		IsRequired:           isRequired,
		AutocompletePosition: autocompletePosition,
		SelectLookup: &apps.Call{
			Path: "/q/event",
			Expand: &apps.Expand{
				OAuth2User: apps.ExpandAll,
				OAuth2App:  apps.ExpandAll,
			},
		},
	}
}

func eventLookup(creq CallRequest) []apps.SelectOption {
	opts := []apps.SelectOption{}

	calService, err := calendar.NewService(creq.ctx, creq.authOption)
	if err != nil {
		creq.log.WithError(err).Warnf("failed to get Calendar client.")
		return nil
	}
	calID := creq.GetValue(fCalendarID, "")
	if calID == "" {
		return nil
	}
	el, err := calService.Events.List(calID).Do()
	if err != nil {
		creq.log.WithError(err).Warnf("failed to get the list of events.")
		return nil
	}

	for _, item := range el.Items {
		opts = append(opts, apps.SelectOption{
			Label: item.Summary,
			Value: item.Id,
		})
	}
	return opts
}
