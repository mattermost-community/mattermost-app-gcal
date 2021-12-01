package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"google.golang.org/api/calendar/v3"
)

func fieldCalendarID(isRequired bool, autocompletePosition int) apps.Field {
	return apps.Field{
		Type:                 apps.FieldTypeDynamicSelect,
		Name:                 fCalendarID,
		Description:          "Choose a Google Calendar",
		IsRequired:           isRequired,
		AutocompletePosition: autocompletePosition,
		SelectDynamicLookup: &apps.Call{
			Path: "/q/cal",
			Expand: &apps.Expand{
				OAuth2App:  apps.ExpandAll,
				OAuth2User: apps.ExpandAll,
			},
		},
	}
}

func calendarIDLookup(creq CallRequest) []apps.SelectOption {
	opts := []apps.SelectOption{}
	calService, err := calendar.NewService(creq.ctx, creq.authOption)
	if err != nil {
		creq.log.WithError(err).Warnf("failed to get Calendar client.")
		return nil
	}

	cl, err := calService.CalendarList.List().Do()
	if err != nil {
		creq.log.WithError(err).Warnf("failed to get the list of Google calendars.")
		return nil
	}

	for _, item := range cl.Items {
		if item.Deleted || item.Hidden {
			continue
		}
		opts = append(opts, apps.SelectOption{
			Label: item.Summary,
			Value: item.Id,
		})
	}
	return opts
}
