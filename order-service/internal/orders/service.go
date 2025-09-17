package orders

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Order struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Service struct {
	httpClient *http.Client
}

func NewService() *Service {
	return &Service{
		httpClient: &http.Client{},
	}
}

func (s *Service) CreateOrder(productID string, quantity int) (*Order, error) {
	resp, err := s.httpClient.Get("http://localhost:3000/products/" + productID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("product not found")
	}

	var product struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}

	// For now, just return a dummy order
	order := &Order{
		ID:        "new-order-id",
		ProductID: productID,
		Quantity:  quantity,
	}

	return order, nil
}