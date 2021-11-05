package function

import (
	"google.golang.org/api/option"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// fieldUseServiceAccount is a predefined command field for
// --use-service-account.
var fieldUseServiceAccount = apps.Field{
	Name: fUseServiceAccount,
	Type: apps.FieldTypeBool,
}

var fieldImpersonateEmail = apps.Field{
	Name: fImpersonateEmail,
	Type: apps.FieldTypeText,
}

func WithCommandAuth(creq CallRequest) option.ClientOption {
	if !creq.BoolValue(fUseServiceAccount) {
		return option.WithTokenSource(creq.ts)
	}

	email := creq.GetValue(fImpersonateEmail, "")
	opt, err := ServiceAccountFromRequest(creq).AuthOption(creq.ctx, email)
	if err != nil {
		// this is used only in the debug command, no reason to make it
		// complicated, just log (and crash?).
		// TODO if used outside the debug, return the error and handle.
		creq.log.WithError(err).Errorf("failed to get service account impersonation auth option")
		return nil
	}
	return opt
}
