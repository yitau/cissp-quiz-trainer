package main

import "context"

// App is the thin Wails-facing application adapter.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// HealthCheck proves the initial Wails binding surface is available.
func (a *App) HealthCheck() string {
	return "ok"
}
