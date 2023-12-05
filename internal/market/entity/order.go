package entity

// type OrderType string

// const (
// 	Buy OrderType = "Buy"
// 	Sell
// )

type Order struct {
	Id            string
	Investor      *Investor
	Asset         *Asset
	Shares        int
	PendingShares int
	Price         float64
	// OrderType     OrderType
	OrderType    string
	Status       string
	Transactions []*Transaction
}

func NewOrder(id string, investor *Investor, asset *Asset, shares int, price float64, orderType string) *Order {
	return &Order{
		Id:            id,
		Investor:      investor,
		Asset:         asset,
		Shares:        shares,
		PendingShares: shares,
		Price:         price,
		OrderType:     orderType,
		Status:        "OPEN",
		Transactions:  []*Transaction{},
	}
}

func (order *Order) CloseOrder() {
	order.Status = "CLOSED"
}
