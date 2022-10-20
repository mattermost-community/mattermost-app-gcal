package function

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var watchList = Command{
	Name:        "list",
	Description: "List personal subscriptions to Google Calendar change notifications",

	Handler: func(creq CallRequest) apps.CallResponse {
		subs, err := creq.ListSubs(creq.Context.ActingUser.Id)
		if err != nil {
			return apps.NewErrorResponse(err)
		}
		if len(subs) == 0 {
			return apps.NewTextResponse("No personal subscriptions.")
		}

		message := "#### List of personal Google Calendar subscriptions.\n\n"
		for _, sub := range subs {
			message += fmt.Sprintf("- **%s** - ID:`%s`, Sync token:`%s`\n\n", sub.CalendarSummary, sub.SubID, sub.SyncToken)
		}

		return RespondWithJSON(creq, message, subs)
	},
}
