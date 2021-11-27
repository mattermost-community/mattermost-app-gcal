package function

import (
	"golang.org/x/oauth2"
)

type User struct {
	Token *oauth2.Token
	Email string
	ID    string
}
