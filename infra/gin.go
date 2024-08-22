package infra

import (
	"golang.org/x/sync/errgroup"
	"net/http"
)

type ServerModel struct {
	Port string
}

type IServerConfig interface {
	Run(server *http.Server) *error
}

func NewServerConfig(model ServerModel) IServerConfig {
	return ServerModel{
		Port: model.Port,
	}
}

func (s ServerModel) Run(server *http.Server) *error {

	var groupRouter errgroup.Group
	groupRouter.Go(func() error {
		return server.ListenAndServe()
	})

	if err := groupRouter.Wait(); err != nil {
		return &err
	}

	return nil
}
