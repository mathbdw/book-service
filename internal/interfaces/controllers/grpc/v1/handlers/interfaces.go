package handlers

import (
	"google.golang.org/grpc"
)

//go:generate mockgen -destination=../../../../../mocks/mock_handler_interfaces.go -package=mocks . ServiceRegistrar

type ServiceRegistrar = grpc.ServiceRegistrar
