package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func fieldCalendarID(isRequired bool, autocompletePosition int) apps.Field {
	return apps.Field{
		Type:                 apps.FieldTypeDynamicSelect,
		Name:                 fCalendarID,
		Description:          "Choose a Google Calendar",
		IsRequired:           isRequired,
		AutocompletePosition: autocompletePosition,
	}
}

func handleCalendarIDLookup(
	filter func(*calendar.CalendarListEntry) bool,
) HandlerFunc {
	h := func(creq CallRequest) []apps.SelectOption {
		opts := []apps.SelectOption{}
		calService, err := calendar.NewService(creq.ctx, option.WithTokenSource(creq.ts))
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
			if filter != nil {
				if !filter(item) {
					continue
				}
			} else {
				if item.Deleted || item.Hidden {
					continue
				}
			}
			opts = append(opts, apps.SelectOption{
				Label: item.Summary,
				Value: item.Id,
			})
		}
		return opts
	}

	return LookupHandler(h)
}
