package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// fieldDebugUseServiceAccount is a predefined command field for
// --use-service-account.
var fieldDebugUseServiceAccount = apps.Field{
	Name: fUseServiceAccount,
	Type: apps.FieldTypeBool,
}

var fieldDebugImpersonateEmail = apps.Field{
	Name: fImpersonateEmail,
	Type: apps.FieldTypeText,
}

var fieldDebugJSON = apps.Field{
	Type:        apps.FieldTypeBool,
	Name:        fJSON,
	Description: "Output JSON",
}

func appendDebugFields(in []apps.Field, creq CallRequest) []apps.Field {
	if !creq.Context.DeveloperMode {
		return in
	}

	return append(in,
		fieldDebugUseServiceAccount,
		fieldDebugImpersonateEmail,
		fieldDebugJSON)
}

func RespondWithJSON(creq CallRequest, message string, v interface{}) apps.CallResponse {
	outJSON := creq.BoolValue(fJSON)
	if outJSON {
		message += "----\n"
		message += utils.JSONBlock(v)
	}
	return apps.NewTextResponse(message)
}
