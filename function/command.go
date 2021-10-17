package function

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type CommandHandler interface {
	Handler
	Metadata() Meta
	Binding(CallRequest) apps.Binding
	Form(CallRequest) apps.Form
}

type Meta struct {
	Name               string
	Path               string
	Hint               string
	Description        string
	RequireAdmin       bool
	RequireGoogleToken bool
}

type command struct {
	Meta

	call    apps.Call
	form    apps.Form
	handler func(CallRequest) apps.CallResponse
}

var _ CommandHandler = (*command)(nil)
var _ Handler = (*command)(nil)

func (c command) path() string {
	if c.Path == "" {
		return "/" + c.Name
	}
	return c.Path
}

func (c command) Metadata() Meta {
	m := c.Meta
	m.Path = c.path()
	return m
}

func (c command) makeCall(base *apps.Call) *apps.Call {
	call := base.Clone()
	if call.Path == "" {
		call.Path = c.path()
	}

	if call.Expand == nil {
		call.Expand = c.call.Expand
	}
	if c.RequireAdmin || c.RequireGoogleToken {
		if call.Expand == nil {
			call.Expand = &apps.Expand{}
		}
		if c.RequireAdmin {
			call.Expand.AdminAccessToken = apps.ExpandAll
		}
		if c.RequireGoogleToken {
			call.Expand.OAuth2User = apps.ExpandAll
		}
	}

	if call.State == nil {
		call.State = c.call.State
	}

	return call
}

func (c command) Form(CallRequest) apps.Form {
	form := *c.form.Clone()
	if form.Icon == "" {
		form.Icon = IconPath
	}

	if form.Call != nil {
		form.Call = c.makeCall(form.Call)
	} else {
		form.Call = c.makeCall(&c.call)
	}

	for i, f := range form.Fields {
		if f.Label == "" {
			form.Fields[i].Label = strings.ReplaceAll(f.Name, "_", "-")
		}
		if f.ModalLabel == "" {
			form.Fields[i].ModalLabel = strings.ReplaceAll(f.Name, "_", " ")
		}
	}

	return form
}

func (c command) Binding(creq CallRequest) apps.Binding {
	form := c.Form(creq)
	call := c.makeCall(&c.call)
	return apps.Binding{
		Location:    apps.Location(c.Name),
		Icon:        IconPath,
		Label:       strings.ReplaceAll(c.Name, "_", "-"),
		Hint:        c.Hint,
		Description: c.Description,
		Call:        call,
		Form:        &form,
	}
}

func (c command) Handle(creq CallRequest) apps.CallResponse {
	if c.handler == nil {
		return apps.NewTextResponse("OK")
	}
	return c.handler(creq)
}
