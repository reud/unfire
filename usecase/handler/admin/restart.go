package admin

import (
	"unfire/usecase"
	"unfire/usecase/batch"
)

type RestartUseCase interface {
	Delete(ctx usecase.RequestContext)
	Reload(ctx usecase.RequestContext)
}

type restartUseCaseImpl struct {
	services batch.Services
}

func NewRestartUseCase(sv batch.Services) RestartUseCase {
	return &restartUseCaseImpl{services: sv}
}

func (ruc *restartUseCaseImpl) Delete(ctx usecase.RequestContext) {
	go ruc.services.DeleteOnce()
}

func (ruc *restartUseCaseImpl) Reload(ctx usecase.RequestContext) {
	go ruc.services.ReloadOnce()
}
