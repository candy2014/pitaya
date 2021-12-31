package services

import (
	"context"
	"github.com/topfreegames/pitaya"

	"github.com/topfreegames/pitaya/component"
)

// Room represents a component that contains a bundle of room related handler
type Room struct {
	component.Base
}

// NewRoom returns a new room
func NewRoom() *Room {
	return &Room{}
}

type LoginRequest struct {
	Account  string
	ServerId int32
}

type LoginResponse struct {
	Account  string
	ServerId int32
}

// Ping returns a pong
func (r *Room) Ping(ctx context.Context) ([]byte, error) {
	return []byte("pong"), nil
}

// Login returns a pong
func (r *Room) Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error) {
	resp := &LoginResponse{
		Account:  "hello world",
		ServerId: request.ServerId,
	}
	return resp, nil
}

func (r *Room) Init() {
	pitaya.AddLogicRoute(1, "room", "room", "login", 1)
}
