package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEstimatePrice(t *testing.T) {
	got, err := estimatePrice(estimateRequest{
		UnitPrice:       25,
		Quantity:        4,
		DiscountPercent: 10,
		TaxPercent:      8.25,
		Shipping:        5,
		Currency:        "usd",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Currency != "USD" || got.Subtotal != 100 || got.Discount != 10 || got.Taxable != 90 || got.Tax != 7.43 || got.GrandTotal != 102.43 {
		t.Fatalf("unexpected estimate: %+v", got)
	}
}

func TestEstimateRejectsInvalidInputs(t *testing.T) {
	cases := []estimateRequest{
		{UnitPrice: -1, Quantity: 1},
		{UnitPrice: 1, Quantity: 0},
		{UnitPrice: 1, Quantity: 1, DiscountPercent: 101},
		{UnitPrice: 1, Quantity: 1, TaxPercent: -1},
		{UnitPrice: 1, Quantity: 1, Shipping: -1},
		{UnitPrice: 1, Quantity: 1, Currency: "US"},
	}
	for i, tc := range cases {
		if _, err := estimatePrice(tc); err == nil {
			t.Fatalf("case %d expected validation error", i)
		}
	}
}

func TestEstimateEndpoint(t *testing.T) {
	body := bytes.NewBufferString(`{"unit_price":12.5,"quantity":2,"discount_percent":0,"tax_percent":10,"shipping":3,"currency":"USD"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/estimate", body)
	rr := httptest.NewRecorder()
	newServer().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var response estimateResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.GrandTotal != 30.5 {
		t.Fatalf("expected grand total 30.5, got %.2f", response.GrandTotal)
	}
}

func TestEstimateEndpointRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/estimate", bytes.NewBufferString(`{"unit_price":10,"quantity":1,"mystery":true}`))
	rr := httptest.NewRecorder()
	newServer().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHealthAndReadiness(t *testing.T) {
	for _, path := range []string{"/healthz", "/readyz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		newServer().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s expected 200, got %d", path, rr.Code)
		}
	}
}
