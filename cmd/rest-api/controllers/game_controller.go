package controllers

import (
	"context"
	"net/http"
)

type GameController struct {
}

func NewGameController() *GameController {
	return &GameController{}
}

func (gc *GameController) Get(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (gc *GameController) Create(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (gc *GameController) Update(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (gc *GameController) Delete(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (gc *GameController) List(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
