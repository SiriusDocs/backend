package services

import (
	// "context"
	"github.com/SiriusDocs/protos/gen/go/fill"
)

type FillService interface {
}

type fillService struct {
	fillClient fill.FillClient
}

func NewFillService(client fill.FillClient) FillService {
	return &fillService{
		fillClient: client,
	}
}
