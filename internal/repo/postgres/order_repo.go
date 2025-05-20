package order

type Repository interface {
	Create()
	List()
}

type repo struct {
	// db *pgxpool.Pool
}

func NewRepository() Repository {
	return &repo{}
}

func (r *repo) Create() {}
func (r *repo) List()   {}
