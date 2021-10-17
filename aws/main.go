package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	function "github.com/mattermost/mattermost-app-gcal/function"
)

func main() {
	function.InitHTTP("")
	lambda.Start(httpadapter.New(http.DefaultServeMux).Proxy)
}
