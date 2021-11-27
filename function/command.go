package function

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Command struct {
	Name        string
	Hint        string
	Description string
	BaseSubmit  *apps.Call
	BaseForm    *apps.Form

	Handler func(CallRequest) apps.CallResponse
}

func (c Command) Path() string {
	return "/" + c.Name
}

func (c Command) Submit(creq CallRequest) *apps.Call {
	if c.BaseSubmit == nil {
		return nil
	}
	s := *c.BaseSubmit.PartialCopy()
	if s.Path == "" {
		s.Path = c.Path()
	}
	return &s
}

func (c Command) Form(creq CallRequest) *apps.Form {
	if c.BaseForm == nil {
		return nil
	}
	f := *c.BaseForm.PartialCopy()
	if f.Icon == "" {
		f.Icon = IconPath
	}
	if f.Submit == nil {
		f.Submit = c.Submit(creq)
	} else if f.Submit.Path == "" {
		f.Submit.Path = c.Path()
	}

	f.Fields = appendDebugFields(f.Fields, creq)
	for i, field := range f.Fields {
		if field.Label == "" {
			f.Fields[i].Label = strings.ReplaceAll(field.Name, "_", "-")
		}
		if field.ModalLabel == "" {
			f.Fields[i].ModalLabel = strings.ReplaceAll(field.Name, "_", " ")
		}
	}

	return &f
}

func (c Command) Binding(creq CallRequest) apps.Binding {
	b := apps.Binding{
		Location:    apps.Location(c.Name),
		Icon:        IconPath,
		// TODO: ticket plugin-apps should do this.
		Label:       strings.ReplaceAll(c.Name, "_", "-"),
		Hint:        c.Hint,
		Description: c.Description,
		Submit:      c.Submit(creq),
		Form:        c.Form(creq),
	}
	return b
}
