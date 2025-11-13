package content

// Core contains shared business logic for content domain
type Core struct {
	service Service
}

func NewCore(service Service) *Core {
	return &Core{
		service: service,
	}
}
