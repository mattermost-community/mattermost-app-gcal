package main

import (
	"fmt"
	"net/http"

	root "github.com/mattermost/mattermost-app-gcal"
	function "github.com/mattermost/mattermost-app-gcal/function"
)

func main() {
	root.InitHTTP("")
	function.InitHTTP("")

	fmt.Println("gcal App started, manifest at http://localhost:4444/manifest.json")
	panic(http.ListenAndServe(":4444", nil))
}
