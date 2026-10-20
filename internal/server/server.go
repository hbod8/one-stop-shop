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
	return api.GetProductsId404JSONResponse{
		Message: api.ProductNotFoundMessage,
	}, nil
}

func (s *Server) PostProducts(ctx context.Context, request api.PostProductsRequestObject) (api.PostProductsResponseObject, error) {
	if request.Body.Id == "" || request.Body.Name == "" || request.Body.Description == "" || request.Body.Price <= 0 {
		return api.PostProducts400JSONResponse{
			Message: api.MissingOrMalformedRequestMessage,
		}, nil
	}
	newProduct := api.Product{
		Id:          request.Body.Id,
		Name:        request.Body.Name,
		Description: request.Body.Description,
		Price:       request.Body.Price,
	}
	products = append(products, newProduct)
	return api.PostProducts201JSONResponse(newProduct), nil
}

func (s *Server) PostAuthLogin(ctx context.Context, request api.PostAuthLoginRequestObject) (api.PostAuthLoginResponseObject, error) {
	if request.Body.Username != nil && *request.Body.Username == "harry" && request.Body.Password != nil && *request.Body.Password == "secret" {
		token := "fake-jwt-token"
		return api.PostAuthLogin200JSONResponse{Token: &token}, nil
	}
	return api.PostAuthLogin401JSONResponse{
		Message: api.InvalidCredentialsMessage,
	}, nil
}
