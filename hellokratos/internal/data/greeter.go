package data

import (
	"context"

	"hellokratos/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

// GreeterRepo is a Greater repo.
type GreeterRepo interface {
	Save(context.Context, *model.Greeter) (*model.Greeter, error)
	Update(context.Context, *model.Greeter) (*model.Greeter, error)
	FindByID(context.Context, int64) (*model.Greeter, error)
	ListByHello(context.Context, string) ([]*model.Greeter, error)
	ListAll(context.Context) ([]*model.Greeter, error)
}

type greeterRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo 创建Greeter仓库实例
func NewGreeterRepo(data *Data, logger log.Logger) GreeterRepo {
	return &greeterRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *greeterRepo) Save(ctx context.Context, g *model.Greeter) (*model.Greeter, error) {
	return g, nil
}

func (r *greeterRepo) Update(ctx context.Context, g *model.Greeter) (*model.Greeter, error) {
	return g, nil
}

func (r *greeterRepo) FindByID(context.Context, int64) (*model.Greeter, error) {
	return nil, nil
}

func (r *greeterRepo) ListByHello(context.Context, string) ([]*model.Greeter, error) {
	return nil, nil
}

func (r *greeterRepo) ListAll(context.Context) ([]*model.Greeter, error) {
	return nil, nil
}
