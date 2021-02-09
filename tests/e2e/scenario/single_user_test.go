package scenario

import (
	"testing"
	"unfire/usecase/handler"
	"unfire/usecase/handler/admin"
)

func TestScenario_SingleUser(t *testing.T) {
	sv := initialize()

	au := handler.NewAuthUseCase()
	ru := admin.NewRestartUseCase(sv)

	u := generateUser()
	u.Work(t, UseCases{
		Au: au,
		Ru: ru,
	})

	initialize()
}
