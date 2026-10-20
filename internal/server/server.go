package server

import (
	"context"

	"onestopshop/api"
)

var _ api.StrictServerInterface = (*Server)(nil)

type Server struct{}

func New() *Server {
	return &Server{}
}

// Hardcoded data for the proof of concept; replace with a store later.
var products = []api.Product{
	{Id: "1", Name: "Widget", Description: "A basic widget", Price: 9.99},
	{Id: "2", Name: "Gadget", Description: "A fancy gadget", Price: 19.99},
}

func (s *Server) GetProducts(ctx context.Context, request api.GetProductsRequestObject) (api.GetProductsResponseObject, error) {
	return api.GetProducts200JSONResponse(products), nil
}

func (s *Server) GetProductsId(ctx context.Context, request api.GetProductsIdRequestObject) (api.GetProductsIdResponseObject, error) {
	for _, p := range products {
		if p.Id == request.Id {
			return api.GetProductsId200JSONResponse(p), nil
		}
	}
	return api.GetProductsId404Response{}, nil
}
