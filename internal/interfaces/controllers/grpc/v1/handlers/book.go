package handlers

import (
	"google.golang.org/grpc"

	"github.com/mathbdw/book/internal/usecases/book"
	pb "github.com/mathbdw/book/proto"
	"github.com/mathbdw/book/internal/interfaces/observability"
)

type BookHandler struct {
	pb.BookServiceServer

	uc     *book.BookUsecases
	observ observability.HandlerObservability
}


func NewBookHandler(app grpc.ServiceRegistrar, uc *book.BookUsecases, observ observability.HandlerObservability) {
	
	handler := &BookHandler{
		observ: observ,
		uc: uc,
	}

	pb.RegisterBookServiceServer(app, handler)
}
