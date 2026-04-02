package shared

type Item struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Price       float64 `json:"price"`
}

type Discount struct {
	Name        string  `json:"name"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	ItemName    string  `json:"item_name"`
}
