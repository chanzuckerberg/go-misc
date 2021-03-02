package setup

import (
	"os"
	"strings"
)

// TODO: Grab from Okta
func GetUsers() ([]string, error) {
	usersList := os.Getenv("CURRENT_USERS") //Comma-delimited users list
	userSlice := strings.Split(usersList, ",")
	return userSlice, nil
}
