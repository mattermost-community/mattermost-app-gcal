module github.com/mattermost/mattermost-app-gcal

go 1.16

require (
	github.com/aws/aws-lambda-go v1.19.1
	github.com/awslabs/aws-lambda-go-api-proxy v0.11.0
	github.com/mattermost/mattermost-plugin-api v0.0.21
	github.com/mattermost/mattermost-plugin-apps v0.7.1-0.20211116123040-849a6caf9a7a
	github.com/mattermost/mattermost-server/v6 v6.0.0-20210906125346-b41b7eae1026
	github.com/pkg/errors v0.9.1
	go.uber.org/zap v1.17.0
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602
	google.golang.org/api v0.44.0
)

replace github.com/mattermost/mattermost-plugin-apps => ../mattermost-plugin-apps
