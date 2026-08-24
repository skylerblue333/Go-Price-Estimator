package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

type estimateRequest struct {
	UnitPrice       float64 `json:"unit_price"`
	Quantity        int     `json:"quantity"`
	DiscountPercent float64 `json:"discount_percent"`
	TaxPercent      float64 `json:"tax_percent"`
	Shipping        float64 `json:"shipping"`
	Currency        string  `json:"currency"`
}

type estimateResponse struct {
	Currency   string  `json:"currency"`
	Subtotal   float64 `json:"subtotal"`
	Discount   float64 `json:"discount"`
	Taxable    float64 `json:"taxable"`
	Tax        float64 `json:"tax"`
	Shipping   float64 `json:"shipping"`
	GrandTotal float64 `json:"grand_total"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

func estimatePrice(req estimateRequest) (estimateResponse, error) {
	if !isFiniteNonNegative(req.UnitPrice) {
		return estimateResponse{}, errors.New("unit_price must be a finite non-negative number")
	}
	if req.Quantity < 1 || req.Quantity > 1_000_000 {
		return estimateResponse{}, errors.New("quantity must be between 1 and 1000000")
	}
	if !isFiniteInRange(req.DiscountPercent, 0, 100) {
		return estimateResponse{}, errors.New("discount_percent must be between 0 and 100")
	}
	if !isFiniteInRange(req.TaxPercent, 0, 100) {
		return estimateResponse{}, errors.New("tax_percent must be between 0 and 100")
	}
	if !isFiniteNonNegative(req.Shipping) {
		return estimateResponse{}, errors.New("shipping must be a finite non-negative number")
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}
	if len(currency) != 3 {
		return estimateResponse{}, errors.New("currency must be a 3-letter code")
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return estimateResponse{}, errors.New("currency must contain only letters")
		}
	}

	subtotal := req.UnitPrice * float64(req.Quantity)
	if math.IsInf(subtotal, 0) || math.IsNaN(subtotal) {
		return estimateResponse{}, errors.New("calculated subtotal is outside supported range")
	}
	discount := subtotal * req.DiscountPercent / 100
	taxable := subtotal - discount
	tax := taxable * req.TaxPercent / 100
	grandTotal := taxable + tax + req.Shipping
	if math.IsInf(grandTotal, 0) || math.IsNaN(grandTotal) {
		return estimateResponse{}, errors.New("calculated total is outside supported range")
	}

	return estimateResponse{
		Currency:   currency,
		Subtotal:   roundMoney(subtotal),
		Discount:   roundMoney(discount),
		Taxable:    roundMoney(taxable),
		Tax:        roundMoney(tax),
		Shipping:   roundMoney(req.Shipping),
		GrandTotal: roundMoney(grandTotal),
	}, nil
}

func isFiniteNonNegative(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func isFiniteInRange(value, min, max float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= min && value <= max
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func handleEstimate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	var req estimateRequest
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON payload"})
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "request body must contain exactly one JSON object"})
		return
	}

	estimate, err := estimatePrice(req)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, estimate)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "sky-price"})
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "service": "sky-price"})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("event=request method=%s path=%s duration_ms=%d", r.Method, r.URL.Path, time.Since(started).Milliseconds())
	})
}

func newServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/estimate", handleEstimate)
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/readyz", handleReady)
	return loggingMiddleware(mux)
}

func main() {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	server := &http.Server{
		Addr:              addr,
		Handler:           newServer(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("event=start service=sky-price addr=%s", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(fmt.Errorf("server failed: %w", err))
	}
}
