package server

import (
	"context"
	"net/http"

	"github.com/EwanGreer/chatatui/internal/repository"
	"github.com/EwanGreer/chatatui/internal/server/api"
)

type ChatServer struct {
	handler    *api.Handler
	srv        *http.Server
	addr       string
	db         *repository.PostgresDB
	onShutdown func()
}

func NewChatServer(h *api.Handler, addr string, db *repository.PostgresDB, onShutdown func()) *ChatServer {
	return &ChatServer{
		handler:    h,
		addr:       addr,
		db:         db,
		onShutdown: onShutdown,
	}
}

func (cs *ChatServer) Start() error {
	cs.srv = &http.Server{
		Addr:    cs.addr,
		Handler: cs.handler.Routes(),
	}

	return cs.srv.ListenAndServe()
}

func (cs *ChatServer) Stop(ctx context.Context) error {
	if cs.onShutdown != nil {
		cs.onShutdown()
	}

	if cs.srv != nil {
		return cs.srv.Shutdown(ctx)
	}

	return nil
}
