package lib

import (
	"encoding/base64"
)

func BasicAuth(user string, pass string) string {
	hash := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	return "Basic " + hash
}
