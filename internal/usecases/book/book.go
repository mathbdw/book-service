package book


type BookUsecases struct {
	Add AddBookUsecase
	Get GetBookUsecase
	List ListBookUsecase
	Remove RemoveBookUsecase
}

// New - constructor 
func New(opts ...BookOptions) *BookUsecases {
	uc := &BookUsecases{}

	for _, opt := range opts{
		opt(uc)
	}

	return uc
}
