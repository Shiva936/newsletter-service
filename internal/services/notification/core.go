package notification

// Core contains shared business logic for subscriber domain
type Core struct {
	service Service
}

func NewCore(service Service) *Core {
	return &Core{
		service: service,
	}
}
