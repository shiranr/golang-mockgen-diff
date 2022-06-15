package example

// mockgen -source=./example/service_example.go -destination=./example/mock/service_example_mock.go

type IService interface {
	DoSomething()
}

type service struct {
}

func NewService() IService {
	return &service{}
}

func (s *service) DoSomething() {
	println("Something")
}
