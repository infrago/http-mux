package http_mux

import (
	"github.com/infrago/http"
	"github.com/infrago/infra"
)

func Driver() http.Driver {
	return &muxDriver{}
}

func init() {
	infra.Register("mux", Driver())
}
