package order

type Service interface {
	Create()
	List()
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) Create() {
	// логика создания
}

func (s *service) List() {
	// логика получения
}
