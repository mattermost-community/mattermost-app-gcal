package root

import (
	"embed"
	_ "embed" // Need to embed manifest file
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
	"github.com/mattermost/mattermost-server/v6/model"
)

//go:embed plugin.json
var pluginManifestData []byte

//go:embed manifest.json
var AppManifestData []byte

//go:embed static
var StaticFS embed.FS

var Manifest model.Manifest
var AppManifest apps.Manifest

func init() {
	_ = json.Unmarshal(pluginManifestData, &Manifest)
	_ = json.Unmarshal(AppManifestData, &AppManifest)
}

func InitHTTP(prefix string) {
	http.HandleFunc(prefix+"/manifest.json",
		httputils.HandleStaticJSONData(AppManifestData))

	http.Handle(prefix+"/static/",
		http.StripPrefix(prefix+"/", http.FileServer(http.FS(StaticFS))))
}
