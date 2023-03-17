package http_default

import (
	"github.com/infrago/http"
)

func Driver() http.Driver {
	return &defaultDriver{}
}

func init() {
	http.Register("default", Driver())
}
