package api

import (
	"encoding/json"
	"net/http"

	"github.com/EwanGreer/chatatui/internal/middleware"
)

type MeHandler struct{}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

func (h *MeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{
		ID:   user.ID.String(),
		Name: user.Name,
	})
}
