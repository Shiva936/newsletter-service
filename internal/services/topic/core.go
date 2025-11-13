package topic

// Core contains shared business logic for topic domain
type Core struct {
	service Service
}

func NewCore(service Service) *Core {
	return &Core{
		service: service,
	}
}
