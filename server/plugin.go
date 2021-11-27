package main

import (
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-server/v6/plugin"

	root "github.com/mattermost/mattermost-app-gcal"
	function "github.com/mattermost/mattermost-app-gcal/function"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin
}

func (p *Plugin) OnActivate() error {
	root.InitHTTP(apps.PluginAppPath)

	function.Log = utils.NewPluginLogger(pluginapi.NewClient(p.API, p.Driver))
	function.AppPathPrefix = apps.PluginAppPath
	function.Init()
	return nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	http.DefaultServeMux.ServeHTTP(w, r)
}
