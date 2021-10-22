package function

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type SimpleCommand struct {
	Name        string
	Hint        string
	Path        string
	Description string

	Submit  apps.Call
	Form    apps.Form
	Binding apps.Binding
	Handler func(CallRequest) apps.CallResponse
}

func (c SimpleCommand) Init() SimpleCommand {
	if c.Path == "" {
		c.Path = "/" + c.Name
	}
	c.Submit = *c.Submit.Clone()
	if c.Submit.Path == "" {
		c.Submit.Path = c.Path
	}

	c.Form = *c.Form.Clone()
	if c.Form.Icon == "" {
		c.Form.Icon = IconPath
	}
	if c.Form.Call == nil {
		c.Form.Call = &c.Submit
	} else if c.Form.Call.Path == "" {
		c.Form.Call.Path = c.Path
	}

	for i, f := range c.Form.Fields {
		if f.Label == "" {
			c.Form.Fields[i].Label = strings.ReplaceAll(f.Name, "_", "-")
		}
		if f.ModalLabel == "" {
			c.Form.Fields[i].ModalLabel = strings.ReplaceAll(f.Name, "_", " ")
		}
	}

	c.Binding = apps.Binding{
		Location:    apps.Location(c.Name),
		Icon:        IconPath,
		Label:       strings.ReplaceAll(c.Name, "_", "-"),
		Hint:        c.Hint,
		Description: c.Description,
		Call:        &c.Submit,
		Form:        &c.Form,
	}

	if c.Handler == nil {
		c.Handler = func(creq CallRequest) apps.CallResponse {
			return apps.NewTextResponse("OK")
		}
	}

	return c
}
