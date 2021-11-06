package function

import (
	"fmt"
	"strings"

	"google.golang.org/api/calendar/v3"
)

func EventDateTimeString(event *calendar.Event) string {
	datetime := ""
	if event.Start.Date != "" {
		// all-day
		datetime = fmt.Sprintf("%s (%s), all day", event.Start.Date, event.Start.TimeZone)
	} else {
		datetime = fmt.Sprintf("%s-%s (%s)", event.Start.DateTime, event.End.DateTime, event.Start.TimeZone)
	}
	if len(event.Recurrence) != 0 {
		datetime += fmt.Sprintf(", %s", event.Recurrence)
	}
	return datetime
}

func EventSummaryString(event *calendar.Event) string {
	return fmt.Sprintf("- **%s**: [%s](%s) - %s.\n",
		EventDateTimeString(event), event.Summary, event.HtmlLink, event.Status)
}

func EventAttendeesString(event *calendar.Event) string {
	if len(event.Attendees) == 0 {
		return ""
	}
	atts := []string{}
	for _, a := range event.Attendees {
		disp := a.DisplayName
		if a.Organizer {
			disp = "**" + disp + "**"
		}
		atts = append(atts,
			fmt.Sprintf("%s (`%s`): %s", disp, a.Email, a.ResponseStatus))
	}
	return strings.Join(atts, ", ")
}
