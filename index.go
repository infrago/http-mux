package http_default

import (
	"github.com/infrago/http"
	"github.com/infrago/infra"
)

func Driver() http.Driver {
	return &defaultDriver{}
}

func init() {
	infra.Register("default", Driver())
}
