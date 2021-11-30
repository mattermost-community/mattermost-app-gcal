package function

import (
	"encoding/base64"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
	"google.golang.org/api/calendar/v3"
)

type Event struct {
	Event *calendar.Event

	RootPostID string
}

func EventDateTimeString(e *calendar.Event) string {
	if e == nil || e.Start == nil || e.End == nil {
		return "(invalid)"
	}
	datetime := ""
	loc, err := time.LoadLocation(e.Start.TimeZone)
	if err != nil {
		return "invalid: " + err.Error()
	}

	if e.Start.Date != "" {
		// all-day
		datetime = fmt.Sprintf("%s (%s), all day", e.Start.Date, loc.String())
	} else {
		start, err := time.ParseInLocation(time.RFC3339, e.Start.DateTime, loc)
		if err != nil {
			return "invalid: " + err.Error()
		}
		end, err := time.ParseInLocation(time.RFC3339, e.End.DateTime, loc)
		if err != nil {
			return "invalid: " + err.Error()
		}
		startFormat := "January 02, 2006 15:04"
		endFormat := "January 02, 2006 15:04"
		if start.Year() == time.Now().Year() {
			startFormat = "January 02, 15:04"
		}
		if start.Year() == end.Year() {
			if start.Day() == end.Day() {
				endFormat = "15:04"
			} else {
				endFormat = "January 02, 15:04"
			}
		}

		dur := end.Sub(start)
		hours := int(dur.Truncate(time.Hour).Hours())
		minutes := int((dur - dur.Truncate(time.Hour)).Minutes())
		durStr := fmt.Sprintf("%vm", minutes)
		if hours > 0 {
			durStr = fmt.Sprintf("%vh, %vm", hours, minutes)
		}
		datetime = fmt.Sprintf("%s-%s, %s (%s)", start.Format(startFormat), end.Format(endFormat), loc.String(), durStr)
	}
	if len(e.Recurrence) != 0 {
		datetime += fmt.Sprintf(", %s", e.Recurrence)
	}
	return datetime
}

func EventSummaryString(e *calendar.Event) string {
	return fmt.Sprintf("[%s](%s)", e.Summary, e.HtmlLink)
}

func EventAttendeesString(e *calendar.Event) string {
	if len(e.Attendees) == 0 {
		return ""
	}
	atts := []string{}
	for _, a := range e.Attendees {
		disp := a.DisplayName
		if a.Organizer {
			disp = "**" + disp + "**"
		}
		atts = append(atts,
			fmt.Sprintf("%s (`%s`): %s", disp, a.Email, a.ResponseStatus))
	}
	return strings.Join(atts, ", ")
}

func EventDiffString(before *Event, after *calendar.Event, calSummary string) string {
	if after.Id == "" {
		return "error: empty ID"
	}
	if before != nil && before.Event.Id != "" && before.Event.Id != after.Id {
		return fmt.Sprintf("ID mismatch: before %q, after %q", before.Event.Id, after.Id)
	}
	if after.Status == "canceled" && before != nil {
		return fmt.Sprintf("**Canceled**: %s", EventSummaryString(before.Event))
	}

	s := ""
	if before == nil {
		s = fmt.Sprintf("###### New in *%s*: %s\n", calSummary, EventSummaryString(after))
	} else {
		s = fmt.Sprintf("Updated: %s\n", EventSummaryString(after))
	}

	if before == nil || utils.ToJSON(before.Event.Start) != utils.ToJSON(before.Event.End) {
		s += fmt.Sprintf("- Time: %s\n", EventDateTimeString(after))
	}
	if before == nil || before.Event.Description != after.Description {
		s += fmt.Sprintf("- Description: %s\n", after.Description)
	}
	if before == nil || before.Event.Status != after.Status {
		s += fmt.Sprintf("- Status: %s\n", after.Status)
	}
	if before == nil || before.Event.Location != after.Location {
		s += fmt.Sprintf("- Location: %s\n", after.Location)
	}

	//TODO: Attendees []*EventAttendee `json:"attendees,omitempty"`
	//TODO: Attachments []*EventAttachment `json:"attachments,omitempty"`
	//TODO: Organizer *EventOrganizer `json:"organizer,omitempty"`

	return s
}

func (creq CallRequest) LoadEvent(googleEmail, calID, eventID string) (*Event, error) {
	var e *Event
	err := appclient.AsBot(creq.Context).KVGet(
		eventKey(googleEmail, calID, eventID), KVEventPrefix, &e)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load event from Mattermost")
	}
	if e == nil {
		return nil, utils.ErrNotFound
	}
	return e, nil
}

func (creq CallRequest) StoreEvent(googleEmail, calID string, e *Event) error {
	_, err := appclient.AsBot(creq.Context).KVSet(
		eventKey(googleEmail, calID, e.Event.Id), KVEventPrefix, e)
	if err != nil {
		return errors.Wrap(err, "failed to save event to Mattermost")
	}
	return nil
}

func (creq CallRequest) DeleteEvent(googleEmail, calID, eventID string) error {
	err := appclient.AsBot(creq.Context).KVDelete(
		eventKey(googleEmail, calID, eventID), KVEventPrefix)
	if err != nil {
		return errors.Wrap(err, "failed to load event from Mattermost")
	}
	return nil
}

func eventKey(googleEmail, calID, eventID string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(path.Join(googleEmail, calID, eventID)))
}
