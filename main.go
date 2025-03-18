package main

import (
	"github.com/alanwade2001/go-sepa-engine-data/repository"
	"github.com/alanwade2001/go-sepa-engine-settle/internal/handler"
	"github.com/alanwade2001/go-sepa-engine-settle/internal/service"
	inf "github.com/alanwade2001/go-sepa-infra"
)

type App struct {
	Infra   *inf.Infra
	Manager *repository.Manager
	Service *service.Execution
	Handler *handler.Execution
}

func NewApp() *App {
	infra := inf.NewInfra()
	manager := repository.NewManager(infra.Persist)
	service := service.NewExecution(manager)
	handler := handler.NewExecution(service, infra.Router)

	app := &App{
		Infra:   infra,
		Manager: manager,
		Service: service,
		Handler: handler,
	}

	return app
}

func (a *App) Run() {
	a.Infra.RunWithTLS()
}

func main() {
	app := NewApp()

	app.Run()

}
