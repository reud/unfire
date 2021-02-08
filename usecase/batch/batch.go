package batch

type BatchService interface {
	Start()
	StartOnce()
}

type ServicesImpl struct {
	Reload BatchService
	Delete BatchService
}

type Services interface {
	Launch()
	LaunchReloadTask()
	ReloadOnce()
	LaunchDeleteTask()
	DeleteOnce()
}

func NewServices(reload BatchService, delete BatchService) Services {
	return &ServicesImpl{
		Reload: reload,
		Delete: delete,
	}
}

func (sv *ServicesImpl) Launch() {
	sv.Delete.Start()
	sv.Reload.Start()
}

func (sv *ServicesImpl) LaunchReloadTask() {
	sv.Reload.Start()
}

func (sv *ServicesImpl) ReloadOnce() {
	sv.Reload.StartOnce()
}

func (sv *ServicesImpl) LaunchDeleteTask() {
	sv.Delete.Start()
}

func (sv *ServicesImpl) DeleteOnce() {
	sv.Delete.StartOnce()
}
