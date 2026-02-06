package clients

import (
	"context"
)

type ClientService struct {
	Q ClientQueries
}

func (s ClientService) GetAll(ctx context.Context) ([]Client, error) {
	return s.Q.ListClients(ctx)
}

func (s ClientService) Create(ctx context.Context, in ClientInput) (int64, error) {
	return s.Q.Insert(ctx, in)
}

func (s ClientService) Delete(ctx context.Context, id int64) (int64, error) {
	return s.Q.Delete(ctx, id)
}

func (s ClientService) GetByID(ctx context.Context, id int64) (Client, error) {
	return s.Q.GetByID(ctx, id)
}

func (s ClientService) Update(ctx context.Context, id int64, in UpdateClientInput) (int64, error) {
	return s.Q.Update(ctx, id, in)
}
