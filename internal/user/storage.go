package user

import "context"

type Storage interface {
	Create(ctx context.Context, user User) (string, error)
	ReadAll(ctx context.Context) ([]User, error)
	Read(ctx context.Context, id string) (User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id string) error
}
