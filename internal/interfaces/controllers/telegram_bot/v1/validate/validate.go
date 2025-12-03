package validate

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
)

// CreateBook - validate message on field book
func CreateBook(mess *tgbotapi.Message) (entities.Book, error) {
	arg := mess.CommandArguments()
	str := strings.Split(arg, "\n")

	if len(str) != 4 {
		return entities.Book{}, errs.New("Invalid book creation format")
	}

	tmp := strings.Trim(str[0], " ")
	tmpTitle := strings.Split(tmp, "Title - ")

	tmp = strings.Trim(str[1], " ")
	tmpDesc := strings.Split(tmp, "Description - ")

	tmp = strings.Trim(str[2], " ")
	tmpYear := strings.Split(tmp, "Year - ")

	tmp = strings.Trim(str[3], " ")
	tmpGenre := strings.Split(tmp, "Genre - ")

	if len(tmpTitle) != 2 || len(tmpDesc) != 2 || len(tmpYear) != 2 || len(tmpGenre) != 2 {
		return entities.Book{}, errs.New("Invalid book creation format")
	}

	year, err := strconv.ParseInt(tmpYear[1], 10, 32)
	if err != nil {
		return entities.Book{}, errs.New("Invalid format year")
	}

	return entities.Book{
		Title:       tmpTitle[1],
		Description: tmpDesc[1],
		Year:        int(year),
		Genre:       tmpGenre[1],
	}, nil
}

func GetBook(mess *tgbotapi.Message) (int64, error) {
	arg := mess.CommandArguments()
	if len(arg) == 0 {
		return 0, errs.New("The argument must not be empty")
	}

	bookId, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return 0, errs.New("The argument must not be empty")
	}

	if bookId < 1 {
		return 0, errs.New("Book not found")
	}

	return bookId, nil
}
