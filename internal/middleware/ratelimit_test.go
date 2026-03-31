package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mocks "github.com/egreerdp/chatatui/internal/middleware/_mocks"
	"github.com/egreerdp/chatatui/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newRateLimiter(cache RateLimitCache, maxReqs int, windowSecs int) *RateLimiter {
	return &RateLimiter{cache: cache, maxReqs: int64(maxReqs), windowSecs: windowSecs}
}

func requestWithUser(userID uuid.UUID) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	user := &repository.User{}
	user.ID = userID
	ctx := contextWithUser(r.Context(), user)
	return r.WithContext(ctx)
}

func contextWithUser(ctx context.Context, user *repository.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimiter_NoUser_PassesThrough(t *testing.T) {
	cache := mocks.NewMockRateLimitCache(t)
	rl := newRateLimiter(cache, 10, 60)

	req := httptest.NewRequest(http.MethodGet, "/", nil) // no user in context
	w := httptest.NewRecorder()

	rl.Middleware(okHandler()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_FirstRequest_SetsExpiry(t *testing.T) {
	userID := uuid.New()
	cache := mocks.NewMockRateLimitCache(t)
	cache.EXPECT().Incr(mock.Anything, userID.String()).Return(int64(1), nil)
	cache.EXPECT().Expire(mock.Anything, userID.String(), 60*time.Second).Return(true, nil)

	rl := newRateLimiter(cache, 10, 60)

	w := httptest.NewRecorder()
	rl.Middleware(okHandler()).ServeHTTP(w, requestWithUser(userID))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_SubsequentRequest_NoExpiry(t *testing.T) {
	userID := uuid.New()
	cache := mocks.NewMockRateLimitCache(t)
	cache.EXPECT().Incr(mock.Anything, userID.String()).Return(int64(5), nil)
	// Expire must NOT be called when count > 1

	rl := newRateLimiter(cache, 10, 60)

	w := httptest.NewRecorder()
	rl.Middleware(okHandler()).ServeHTTP(w, requestWithUser(userID))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_AtLimit_Allowed(t *testing.T) {
	userID := uuid.New()
	cache := mocks.NewMockRateLimitCache(t)
	cache.EXPECT().Incr(mock.Anything, userID.String()).Return(int64(10), nil)

	rl := newRateLimiter(cache, 10, 60)

	w := httptest.NewRecorder()
	rl.Middleware(okHandler()).ServeHTTP(w, requestWithUser(userID))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiter_ExceedsLimit_Returns429(t *testing.T) {
	userID := uuid.New()
	cache := mocks.NewMockRateLimitCache(t)
	cache.EXPECT().Incr(mock.Anything, userID.String()).Return(int64(11), nil)

	rl := newRateLimiter(cache, 10, 60)

	w := httptest.NewRecorder()
	rl.Middleware(okHandler()).ServeHTTP(w, requestWithUser(userID))

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "60", w.Header().Get("Retry-After"))
}

func TestRateLimiter_IncrError_Returns500(t *testing.T) {
	userID := uuid.New()
	cache := mocks.NewMockRateLimitCache(t)
	cache.EXPECT().Incr(mock.Anything, userID.String()).Return(int64(0), errors.New("redis unavailable"))

	rl := newRateLimiter(cache, 10, 60)

	w := httptest.NewRecorder()
	rl.Middleware(okHandler()).ServeHTTP(w, requestWithUser(userID))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRateLimiter_ExpireError_Returns500(t *testing.T) {
	userID := uuid.New()
	cache := mocks.NewMockRateLimitCache(t)
	cache.EXPECT().Incr(mock.Anything, userID.String()).Return(int64(1), nil)
	cache.EXPECT().Expire(mock.Anything, userID.String(), 60*time.Second).Return(false, errors.New("redis unavailable"))

	rl := newRateLimiter(cache, 10, 60)

	w := httptest.NewRecorder()
	rl.Middleware(okHandler()).ServeHTTP(w, requestWithUser(userID))

	require.Equal(t, http.StatusInternalServerError, w.Code)
}
