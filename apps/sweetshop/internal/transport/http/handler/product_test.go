//go:build testing

package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	coretesting "github.com/bbsbb/go-edge/core/testing"
)

type ProductSuite struct {
	IntegrationSuite
}

func (s *ProductSuite) TestCreateProduct() {
	resp := s.CreateProduct("Vanilla Scoop", "ice_cream", 350)

	s.Assert().NotEmpty(resp["id"])
	s.Assert().Equal("Vanilla Scoop", resp["name"])
	s.Assert().Equal("ice_cream", resp["category"])
	s.Assert().Equal(float64(350), resp["price_cents"])
}

func (s *ProductSuite) TestCreateProduct_ValidationErrors() {
	cases := []struct {
		name     string
		body     map[string]any
		wantCode int
	}{
		{"invalid category", map[string]any{"name": "Bad", "category": "candy", "price_cents": 100}, http.StatusBadRequest},
		{"zero price", map[string]any{"name": "Free", "category": "ice_cream", "price_cents": 0}, http.StatusBadRequest},
		{"negative price", map[string]any{"name": "Neg", "category": "ice_cream", "price_cents": -1}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			req := coretesting.JSONRequest(s.T(), http.MethodPost, "/products", tc.body)
			rec := s.Do(req)
			s.Assert().Equal(tc.wantCode, rec.Code)
		})
	}
}

func (s *ProductSuite) TestGetProduct() {
	created := s.CreateProduct("Chocolate Scoop", "ice_cream", 400)

	req := httptest.NewRequest(http.MethodGet, "/products/"+created["id"].(string), nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	var resp map[string]any
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Equal(created["id"], resp["id"])
	s.Assert().Equal("Chocolate Scoop", resp["name"])
}

func (s *ProductSuite) TestGetProduct_NotFound() {
	req := httptest.NewRequest(http.MethodGet, "/products/019505e0-0000-7000-8000-000000000000", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusNotFound, rec.Code)
}

func (s *ProductSuite) TestGetProduct_InvalidID() {
	req := httptest.NewRequest(http.MethodGet, "/products/not-a-uuid", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ProductSuite) TestListProducts() {
	s.CreateProduct("Vanilla", "ice_cream", 350)
	s.CreateProduct("Fluffy", "marshmallow", 200)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	var resp []map[string]any
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Len(resp, 2)
}

func (s *ProductSuite) TestListProducts_Empty() {
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	var resp []map[string]any
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Empty(resp)
}

func (s *ProductSuite) TestUpdateProduct() {
	created := s.CreateProduct("Old Name", "ice_cream", 300)

	req := coretesting.JSONRequest(s.T(), http.MethodPut, "/products/"+created["id"].(string), map[string]any{
		"name": "New Name", "category": "marshmallow", "price_cents": 500,
	})
	rec := s.Do(req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	var resp map[string]any
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Equal("New Name", resp["name"])
	s.Assert().Equal("marshmallow", resp["category"])
	s.Assert().Equal(float64(500), resp["price_cents"])
}

func (s *ProductSuite) TestUpdateProduct_NotFound() {
	req := coretesting.JSONRequest(s.T(), http.MethodPut, "/products/019505e0-0000-7000-8000-000000000000", map[string]any{
		"name": "X", "category": "ice_cream", "price_cents": 100,
	})
	rec := s.Do(req)

	s.Assert().Equal(http.StatusNotFound, rec.Code)
}

func (s *ProductSuite) TestUpdateProduct_InvalidID() {
	req := coretesting.JSONRequest(s.T(), http.MethodPut, "/products/not-a-uuid", map[string]any{
		"name": "X", "category": "ice_cream", "price_cents": 100,
	})
	rec := s.Do(req)

	s.Assert().Equal(http.StatusBadRequest, rec.Code)
}

func (s *ProductSuite) TestDeleteProduct() {
	created := s.CreateProduct("To Delete", "ice_cream", 100)

	req := httptest.NewRequest(http.MethodDelete, "/products/"+created["id"].(string), nil)
	rec := s.Do(req)
	s.Assert().Equal(http.StatusNoContent, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/products/"+created["id"].(string), nil)
	rec = s.Do(req)
	s.Assert().Equal(http.StatusNotFound, rec.Code)
}

func (s *ProductSuite) TestDeleteProduct_NotFound() {
	req := httptest.NewRequest(http.MethodDelete, "/products/019505e0-0000-7000-8000-000000000000", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusNotFound, rec.Code)
}

func (s *ProductSuite) TestDeleteProduct_InvalidID() {
	req := httptest.NewRequest(http.MethodDelete, "/products/not-a-uuid", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusBadRequest, rec.Code)
}

func TestProductSuite(t *testing.T) {
	suite.Run(t, new(ProductSuite))
}
