package function

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var watchList = Command{
	Name:        "list",
	Description: "List personal subscriptions to Google Calendar change notifications",

	Handler: func(creq CallRequest) apps.CallResponse {
		subs, err := creq.ListSubs(creq.Context.ActingUserID)
		if err != nil {
			return apps.NewErrorResponse(err)
		}
		if len(subs) == 0 {
			return apps.NewTextResponse("No personal subscriptions.")
		}

		message := "#### List of personal Google Calendar subscriptions."
		for _, sub := range subs {
			message += fmt.Sprintf("- %s\n  ID: `%s`, Sync token: `%s`\n\n", sub.CalendarID, sub.SubID, sub.SyncToken)
		}

		outJSON := creq.BoolValue(fJSON)
		if outJSON {
			message += "----\n"
			message += utils.JSONBlock(subs)
		}

		return apps.NewTextResponse(message)
	},
}
