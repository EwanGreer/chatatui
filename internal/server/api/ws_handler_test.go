package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mocks "github.com/egreerdp/chatatui/internal/server/api/_mocks"
	"github.com/egreerdp/chatatui/internal/server/hub"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newWSHandlerRouter wires a WSHandler into a chi router matching the real route
// shape, without starting the hub status goroutine.
func newWSHandlerRouter(svc ChatService) http.Handler {
	h := &WSHandler{
		hub:                 hub.NewHub(),
		svc:                 svc,
		messageHistoryLimit: 50,
	}
	r := chi.NewRouter()
	r.Get("/ws/{roomID}", h.Handle)
	return r
}

func parseErrorResponse(t *testing.T, body []byte) apiError {
	t.Helper()
	var resp apiError
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func TestWSHandler_MissingRoomID(t *testing.T) {
	svc := mocks.NewMockChatService(t)

	// Register a route without the {roomID} segment so chi sets it to "".
	h := &WSHandler{hub: hub.NewHub(), svc: svc, messageHistoryLimit: 50}
	r := chi.NewRouter()
	r.Get("/ws/", h.Handle)

	req := httptest.NewRequest(http.MethodGet, "/ws/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseErrorResponse(t, w.Body.Bytes())
	assert.Equal(t, "ROOM_REQUIRED", resp.Code)
}

func TestWSHandler_InvalidRoomUUID(t *testing.T) {
	svc := mocks.NewMockChatService(t)
	router := newWSHandlerRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/ws/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseErrorResponse(t, w.Body.Bytes())
	assert.Equal(t, "INVALID_ROOM_ID", resp.Code)
}

func TestWSHandler_RoomNotFound(t *testing.T) {
	roomID := uuid.New()
	svc := mocks.NewMockChatService(t)
	svc.EXPECT().GetRoom(roomID).Return(nil, gorm.ErrRecordNotFound)

	router := newWSHandlerRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/ws/"+roomID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseErrorResponse(t, w.Body.Bytes())
	assert.Equal(t, "ROOM_NOT_FOUND", resp.Code)
}

func TestWSHandler_RoomLookupInternalError(t *testing.T) {
	roomID := uuid.New()
	svc := mocks.NewMockChatService(t)
	svc.EXPECT().GetRoom(roomID).Return(nil, errors.New("db connection lost"))

	router := newWSHandlerRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/ws/"+roomID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resp := parseErrorResponse(t, w.Body.Bytes())
	assert.Equal(t, "INTERNAL_ERROR", resp.Code)
}
