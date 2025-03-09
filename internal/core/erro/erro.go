package erro

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	InvalidToken   	= status.Errorf(codes.Unauthenticated, "invalid beard token")
	MissingData 	= status.Errorf(codes.InvalidArgument, "missing metadata")
)