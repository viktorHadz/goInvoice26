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

func (s ClientService) Create(ctx context.Context, in CreateClientInput) (int64, error) {
	return s.Q.Insert(ctx, in)
}
