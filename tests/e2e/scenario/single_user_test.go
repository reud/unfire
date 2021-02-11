package scenario

import (
	"testing"
	"time"
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
	// wait for goroutine
	time.Sleep(2 * time.Second)

	sv.LaunchDeleteTask()
	sv.LaunchReloadTask()

	initialize()
}

func TestScenario_TenUser(t *testing.T) {
	sv := initialize()

	au := handler.NewAuthUseCase()
	ru := admin.NewRestartUseCase(sv)

	i := 0
	for i < 10 {
		u := generateUser()
		u.Work(t, UseCases{
			Au: au,
			Ru: ru,
		})
		i++
	}

	// wait for goroutine
	time.Sleep(5 * time.Second)

	sv.LaunchDeleteTask()
	sv.LaunchReloadTask()

	initialize()
}
