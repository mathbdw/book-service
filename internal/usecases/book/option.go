package book

// BookOption -.
type BookOptions func(*BookUsecases)

// WithAddBookUsecase - Set usecase add_book
func WithAddBookUsecase(uc AddBookUsecase) BookOptions {
	return func(b *BookUsecases) {
		b.Add = uc
	}
}

// WithGetBookUsecase - Set usecase get_book
func WithGetBookUsecase(uc GetBookUsecase) BookOptions {
	return func(b *BookUsecases) {
		b.Get = uc
	}
}

// WithListBookUsecase - Set usecase list_book
func WithListBookUsecase(uc ListBookUsecase) BookOptions {
	return func(b *BookUsecases) {
		b.List = uc
	}
}

// WithRemoveBookUsecase - Set usecase remove_book
func WithRemoveBookUsecase(uc RemoveBookUsecase) BookOptions {
	return func(b *BookUsecases) {
		b.Remove = uc
	}
}
