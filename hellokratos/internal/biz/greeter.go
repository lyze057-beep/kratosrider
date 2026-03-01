package biz

import (
	"context"

	"hellokratos/internal/data"
	"hellokratos/internal/data/model"

	v1 "hellokratos/api/helloworld/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	// ErrUserNotFound is user not found.
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// GreeterUsecase is a Greeter usecase.
type GreeterUsecase struct {
	repo data.GreeterRepo
}

// NewGreeterUsecase new a Greeter usecase.
func NewGreeterUsecase(repo data.GreeterRepo) *GreeterUsecase {
	return &GreeterUsecase{repo: repo}
}

// CreateGreeter creates a Greeter, and returns the new Greeter.
func (uc *GreeterUsecase) CreateGreeter(ctx context.Context, g *model.Greeter) (*model.Greeter, error) {
	log.Infof("CreateGreeter: %v", g.Hello)
	return uc.repo.Save(ctx, g)
}
