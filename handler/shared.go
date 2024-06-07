package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
)

var UploadDir string
var Protocol string
var Hostname string
var Port int

func render(w http.ResponseWriter, r *http.Request, component templ.Component) error {
	return component.Render(r.Context(), w)
}

func MakeHandler(h func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			slog.Error("internal server error", "err", err, "path", r.URL.Path)
		}
	}
}

func URLBase(protocol, hostname string, port int) string {
	return fmt.Sprintf("%s://%s:%d", protocol, hostname, port)
}
