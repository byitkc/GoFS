package handler

import (
	"net/http"

	"github.com/byitkc/GoFS/view/settings"
)

func HandleSettingsIndex(w http.ResponseWriter, r *http.Request) error {
	return settings.Index().Render(r.Context(), w)
}
