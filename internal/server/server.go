package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"onestopshop/api"
)

var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	auth *Auth
}

func New(auth *Auth) *Server {
	return &Server{
		auth: auth,
	}
}

// Hardcoded data for the proof of concept; replace with a store later.
var products = []api.Product{
	{Id: "1", Name: "Widget", Description: "A basic widget", Price: 999},
	{Id: "2", Name: "Gadget", Description: "A fancy gadget", Price: 1999},
}

var nextProductID = 3

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	strict := api.NewStrictHandlerWithOptions(s, nil, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  s.HandleRequestError,
		ResponseErrorHandlerFunc: s.HandleResponseError,
	})
	api.HandlerFromMux(strict, mux)
	return mux
}

func (s *Server) HandleRequestError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
		netErr    net.Error
	)

	var msg api.BadRequestErrorMessage

	switch {
	case errors.Is(err, io.EOF):
		msg = api.MissingRequestBadRequestMessage
	case errors.Is(err, io.ErrUnexpectedEOF):
		msg = api.UnexpectedEndOfRequestBadRequestMessage
	case errors.As(err, &syntaxErr):
		msg = api.MalformedRequestBadRequestMessage
	case errors.As(err, &typeErr):
		msg = api.InvalidTypesInRequestBadRequestMessage
	case errors.As(err, &netErr):
		msg = api.RequestTimedOutBadRequestMessage
	default:
		msg = api.MalformedRequestBadRequestMessage
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(api.BadRequestError{Message: msg})
}

func (s *Server) HandleResponseError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
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
		ProductNotFoundErrorJSONResponse: api.ProductNotFoundErrorJSONResponse{
			Message: api.ProductNotFoundErrorMessageProductNotFoundErrorMessage,
		},
	}, nil
}

func (s *Server) PostProducts(ctx context.Context, request api.PostProductsRequestObject) (api.PostProductsResponseObject, error) {
	if request.Body.Name == "" || request.Body.Description == "" || request.Body.Price <= 0 {
		return api.PostProducts400JSONResponse{
			BadRequestErrorJSONResponse: MissingBadRequestErrorJSON(),
		}, nil
	}
	newProduct := api.Product{
		Id:          GetNextProductID(),
		Name:        request.Body.Name,
		Description: request.Body.Description,
		Price:       request.Body.Price,
	}
	products = append(products, newProduct)
	return api.PostProducts201JSONResponse(newProduct), nil
}

func GetNextProductID() string {
	nextProductID++
	return fmt.Sprintf("%d", nextProductID-1)
}

func (s *Server) PostAuthLogin(ctx context.Context, request api.PostAuthLoginRequestObject) (api.PostAuthLoginResponseObject, error) {
	if request.Body.Username == nil || *request.Body.Username == "" || request.Body.Password == nil || *request.Body.Password == "" {
		return api.PostAuthLogin400JSONResponse{
			BadRequestErrorJSONResponse: MissingBadRequestErrorJSON(),
		}, nil
	}
	token, refreshToken, err := s.auth.authenticate(*request.Body.Username, *request.Body.Password)
	if err == ErrInvalidCredentials {
		return api.PostAuthLogin401JSONResponse{
			Message: api.InvalidCredentialsMessage,
		}, nil
	} else {
		return api.PostAuthLogin200JSONResponse{Token: token, RefreshToken: refreshToken}, nil
	}
}

func MissingBadRequestErrorJSON() api.BadRequestErrorJSONResponse {
	return api.BadRequestErrorJSONResponse{Message: api.MissingRequestBadRequestMessage}
}
