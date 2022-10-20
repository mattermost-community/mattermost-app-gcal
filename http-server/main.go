package main

import (
	"fmt"
	"net/http"
	"time"

	root "github.com/mattermost/mattermost-app-gcal"
	function "github.com/mattermost/mattermost-app-gcal/function"
)

func main() {
	root.InitHTTP("")

	function.AppPathPrefix = ""
	function.Init()

	listen := ":4444"
	server := &http.Server{
		Addr:              listen,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	fmt.Printf("gcal App started, manifest at http://localhost%s/manifest.json\n", listen)
	panic(server.ListenAndServe())
}
