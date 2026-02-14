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

type OrderSuite struct {
	IntegrationSuite
}

func (s *OrderSuite) TestOpenOrder() {
	resp := s.OpenOrder()

	s.Assert().NotEmpty(resp["id"])
	s.Assert().Equal("open", resp["status"])
	s.Assert().Equal(float64(0), resp["total_cents"])
}

func (s *OrderSuite) TestGetOrder() {
	created := s.OpenOrder()

	req := httptest.NewRequest(http.MethodGet, "/orders/"+created["id"].(string), nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	var resp map[string]any
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Equal(created["id"], resp["id"])
	s.Assert().Equal("open", resp["status"])
}

func (s *OrderSuite) TestGetOrder_NotFound() {
	req := httptest.NewRequest(http.MethodGet, "/orders/019505e0-0000-7000-8000-000000000000", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusNotFound, rec.Code)
}

func (s *OrderSuite) TestGetOrder_InvalidID() {
	req := httptest.NewRequest(http.MethodGet, "/orders/not-a-uuid", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusBadRequest, rec.Code)
}

func (s *OrderSuite) TestAddItem() {
	product := s.CreateProduct("Vanilla", "ice_cream", 350)
	order := s.OpenOrder()

	req := coretesting.JSONRequest(s.T(), http.MethodPost, "/orders/"+order["id"].(string)+"/items", map[string]any{
		"product_id": product["id"], "quantity": 2,
	})
	rec := s.Do(req)

	s.Assert().Equal(http.StatusCreated, rec.Code)
	var resp map[string]any
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Equal(product["id"], resp["product_id"])
	s.Assert().Equal(float64(2), resp["quantity"])
	s.Assert().Equal(float64(350), resp["price_cents"])
	s.Assert().Equal(float64(700), resp["line_total_cents"])
}

func (s *OrderSuite) TestAddItem_OrderWithTotal() {
	product1 := s.CreateProduct("Vanilla", "ice_cream", 350)
	product2 := s.CreateProduct("Fluffy", "marshmallow", 200)
	order := s.OpenOrder()

	for _, item := range []struct {
		productID string
		quantity  int
	}{
		{product1["id"].(string), 2},
		{product2["id"].(string), 3},
	} {
		req := coretesting.JSONRequest(s.T(), http.MethodPost, "/orders/"+order["id"].(string)+"/items", map[string]any{
			"product_id": item.productID, "quantity": item.quantity,
		})
		rec := s.Do(req)
		s.Require().Equal(http.StatusCreated, rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/orders/"+order["id"].(string), nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	var resp map[string]any
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Equal(float64(1300), resp["total_cents"])
	items := resp["items"].([]any)
	s.Assert().Len(items, 2)
}

func (s *OrderSuite) TestAddItem_ValidationErrors() {
	order := s.OpenOrder()

	cases := []struct {
		name     string
		body     map[string]any
		wantCode int
	}{
		{"invalid product ID", map[string]any{"product_id": "not-a-uuid", "quantity": 1}, http.StatusBadRequest},
		{"product not found", map[string]any{"product_id": "019505e0-0000-7000-8000-000000000000", "quantity": 1}, http.StatusNotFound},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			req := coretesting.JSONRequest(s.T(), http.MethodPost, "/orders/"+order["id"].(string)+"/items", tc.body)
			rec := s.Do(req)
			s.Assert().Equal(tc.wantCode, rec.Code)
		})
	}
}

func (s *OrderSuite) TestAddItem_ClosedOrder() {
	product := s.CreateProduct("Vanilla", "ice_cream", 350)
	order := s.OpenOrder()

	closeReq := httptest.NewRequest(http.MethodPost, "/orders/"+order["id"].(string)+"/close", nil)
	closeRec := s.Do(closeReq)
	s.Require().Equal(http.StatusOK, closeRec.Code)

	req := coretesting.JSONRequest(s.T(), http.MethodPost, "/orders/"+order["id"].(string)+"/items", map[string]any{
		"product_id": product["id"], "quantity": 1,
	})
	rec := s.Do(req)

	s.Assert().Equal(http.StatusUnprocessableEntity, rec.Code)
}

func (s *OrderSuite) TestCloseOrder() {
	order := s.OpenOrder()

	req := httptest.NewRequest(http.MethodPost, "/orders/"+order["id"].(string)+"/close", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	var resp map[string]any
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Equal("closed", resp["status"])
}

func (s *OrderSuite) TestCloseOrder_AlreadyClosed() {
	order := s.OpenOrder()

	req := httptest.NewRequest(http.MethodPost, "/orders/"+order["id"].(string)+"/close", nil)
	rec := s.Do(req)
	s.Require().Equal(http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/orders/"+order["id"].(string)+"/close", nil)
	rec = s.Do(req)

	s.Assert().Equal(http.StatusUnprocessableEntity, rec.Code)
}

func (s *OrderSuite) TestCloseOrder_NotFound() {
	req := httptest.NewRequest(http.MethodPost, "/orders/019505e0-0000-7000-8000-000000000000/close", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusNotFound, rec.Code)
}

func (s *OrderSuite) TestCloseOrder_InvalidID() {
	req := httptest.NewRequest(http.MethodPost, "/orders/bad-id/close", nil)
	rec := s.Do(req)

	s.Assert().Equal(http.StatusBadRequest, rec.Code)
}

func TestOrderSuite(t *testing.T) {
	suite.Run(t, new(OrderSuite))
}
