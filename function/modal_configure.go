package function

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
)

func configureModal(creq CallRequest) apps.CallResponse {
	clientID := creq.GetValue(fClientID, "")
	clientSecret := creq.GetValue(fClientSecret, "")
	mode := creq.GetValue(fMode, "")
	apiKey := creq.GetValue(fAPIKey, "")
	serviceAccount := creq.GetValue(fAccountJSON, "")

	asActingUser := appclient.AsActingUser(creq.Context)
	sa := NewServiceAccount(mode, apiKey, serviceAccount)
	err := asActingUser.StoreOAuth2App(apps.OAuth2App{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Data:         sa,
	})
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to store Oauth2 configuration to Mattermost"))
	}

	return apps.NewTextResponse(
		"updated OAuth client credentials:\n"+
			"  - Client ID, ending in %q\n"+
			"  - Client Secret, ending in %q\n"+
			"  - %s\n",
		utils.LastN(clientID, 8),
		utils.LastN(clientSecret, 4),
		sa.String(),
	)
}

func configureModalForm(creq CallRequest) (apps.Form, error) {
	clientID := creq.GetValue(fClientID, creq.Context.OAuth2.ClientID)
	clientSecret := creq.GetValue(fClientSecret, creq.Context.OAuth2.ClientSecret)

	prevServiceAccount := ServiceAccountFromRequest(creq)
	mode := creq.GetValue(fMode, prevServiceAccount.Mode)
	apiKey := creq.GetValue(fAPIKey, prevServiceAccount.APIKey)
	accountJSON := creq.GetValue(fAccountJSON, prevServiceAccount.AccountJSON)

	f := apps.Form{
		Submit: &apps.Call{
			Path: "/configure-modal",
			Expand: &apps.Expand{
				ActingUserAccessToken: apps.ExpandAll,
				ActingUser:            apps.ExpandSummary,
			},
		},
		Source: &apps.Call{
			Path: "/f/configure-modal",
			Expand: &apps.Expand{
				ActingUserAccessToken: apps.ExpandAll,
				ActingUser:            apps.ExpandSummary,
			},
		},
		Title: "Configure Google Calendar App credentials",
		Header: "" +
			"- Client ID and Secret are needed for the Mattermost users to authenticate to Google. " +
			"  [documentation](https://TODO)" +
			"\n" +
			"- Service Account is needed for to process incoming webhooks from Google. " +
			"  [documentation](https://TODO)" +
			"\n\n",
		Icon: IconPath,
		Fields: []apps.Field{
			{
				Name:        fClientID,
				ModalLabel:  "Client ID",
				Type:        apps.FieldTypeText,
				Description: "Google Client ID for the Mattermost App.",
				IsRequired:  true,
				Value:       clientID,
			},
			{
				Type:        apps.FieldTypeText,
				TextSubtype: apps.TextFieldSubtypePassword,
				Name:        fClientSecret,
				ModalLabel:  "Client Secret",
				Description: "Google Client Secret for the Mattermost App.",
				IsRequired:  true,
				Value:       clientSecret,
			},
		},
	}

	serviceAccountModes := map[string]string{
		"":           "Do not use a Service Account",
		fAPIKey:      "Use a Google API Key",
		fAccountJSON: "Use a Google Service Account",
	}
	defValue := apps.SelectOption{
		Value: mode,
		Label: serviceAccountModes[mode],
	}
	field := apps.Field{
		Type:          apps.FieldTypeStaticSelect,
		Name:          fMode,
		ModalLabel:    "Service Account",
		Description:   "What kind of Google service account to use to process incoming change notifications.",
		IsRequired:    true,
		SelectRefresh: true,
	}
	for v, l := range serviceAccountModes {
		field.SelectStaticOptions = append(field.SelectStaticOptions, apps.SelectOption{
			Label: l,
			Value: v,
		})
		if v == mode {
			field.Value = defValue
		}
	}
	f.Fields = append(f.Fields, field)

	switch mode {
	case fAPIKey:
		f.Fields = append(f.Fields, apps.Field{
			Type:        apps.FieldTypeText,
			TextSubtype: apps.TextFieldSubtypePassword,
			Name:        fAPIKey,
			ModalLabel:  "API Key",
			Description: "Google API Key for the Mattermost App, no need if using the serv.",
			IsRequired:  true,
			Value:       apiKey,
		})

	case fAccountJSON:
		f.Fields = append(f.Fields, apps.Field{
			Type:          apps.FieldTypeText,
			TextSubtype:   apps.TextFieldSubtypeTextarea,
			TextMaxLength: 32 * 1024,
			Name:          fAccountJSON,
			ModalLabel:    "Service Account (JSON)",
			Description:   "Google Service Account for the Mattermost App. Please open the downloaded credentials JSON file, and paste its contents here.",
			IsRequired:    true,
			Value:         accountJSON,
		})
	}

	return f, nil
}
