package entity

import (
	"container/heap"
	"sync"
)

type Book struct {
	Order         []*Order
	Transactions  []*Transaction
	OrdersChan    chan *Order
	OrdersChanOut chan *Order
	Wg            *sync.WaitGroup
}

func NewBook(ordersChanIn chan *Order, ordersChanOut chan *Order, wg *sync.WaitGroup) *Book {
	return &Book{
		Order:         []*Order{},
		Transactions:  []*Transaction{},
		OrdersChan:    ordersChanIn,
		OrdersChanOut: ordersChanOut,
		Wg:            wg,
	}
}

func (b *Book) Trade() {
	// buyOrders := NewOrderQueue()
	// sellOrders := NewOrderQueue()

	buyOrders := make(map[string]*OrderQueue)
	sellOrders := make(map[string]*OrderQueue)

	// heap.Init(buyOrders)
	// heap.Init(sellOrders)

	for order := range b.OrdersChan {
		asset := order.Asset.ID

		if buyOrders[asset] == nil {
			buyOrders[asset] = NewOrderQueue()
			heap.Init(buyOrders[asset])
		}

		if sellOrders[asset] == nil {
			sellOrders[asset] = NewOrderQueue()
			heap.Init(sellOrders[asset])
		}

		if order.OrderType == "BUY" {
			buyOrders[asset].Push(order)
			if sellOrders[asset].Len() > 0 && sellOrders[asset].Orders[0].Price <= order.Price {
				b.buyOrder(sellOrders[asset], order)
			}
		} else if order.OrderType == "SELL" {
			sellOrders[asset].Push(order)
			if sellOrders[asset].Len() > 0 && sellOrders[asset].Orders[0].Price <= order.Price {
				b.sellOrder(buyOrders[asset], order)
			}
		}
	}
}

func (b *Book) buyOrder(sellOrders *OrderQueue, order *Order) {
	sellOrder := sellOrders.Pop().(*Order)
	if sellOrder.PendingShares > 0 {
		//TODO: Precisa fazer lógica do minShares e colocar no lugar de order.Shares
		transaction := NewTransaction(sellOrder, order, order.Shares, order.Price)
		b.addTransaction(transaction, b.Wg)
		sellOrder.Transactions = append(sellOrder.Transactions, transaction)
		order.Transactions = append(order.Transactions, transaction)
		b.OrdersChanOut <- sellOrder
		b.OrdersChanOut <- order
		if sellOrder.PendingShares > 0 {
			sellOrders.Push(sellOrder)
		}
	}
}

func (b *Book) sellOrder(buyOrders *OrderQueue, order *Order) {
	if buyOrders.Len() > 0 && buyOrders.Orders[0].Price >= order.Price {
		buyOrder := buyOrders.Pop().(*Order)
		if buyOrder.PendingShares > 0 {
			//TODO: Precisa fazer lógica do minShares e colocar no lugar de order.Shares
			transaction := NewTransaction(order, buyOrder, order.Shares, order.Price)
			b.addTransaction(transaction, b.Wg)
			buyOrder.Transactions = append(buyOrder.Transactions, transaction)
			order.Transactions = append(order.Transactions, transaction)
			b.OrdersChanOut <- buyOrder
			b.OrdersChanOut <- order
			if buyOrder.PendingShares > 0 {
				buyOrders.Push(buyOrder)
			}
		}
	}

}

func (b *Book) addTransaction(transaction *Transaction, wg *sync.WaitGroup) {
	defer wg.Done()

	sellingOrder := transaction.SellingOrder
	sellingShares := sellingOrder.PendingShares
	buyingOrder := transaction.BuyingOrder
	buyingShares := buyingOrder.PendingShares

	minShares := sellingShares

	if buyingShares < minShares {
		minShares = buyingShares
	}

	sellingOrder.Investor.UpdateAssetPosition(sellingOrder.Asset.ID, -minShares)
	sellingOrder.PendingShares -= minShares

	buyingOrder.Investor.UpdateAssetPosition(buyingOrder.Asset.ID, minShares)
	buyingOrder.PendingShares -= minShares

	transaction.CalculateTotal(transaction.Shares, buyingOrder.Price)

	if buyingOrder.PendingShares == 0 {
		buyingOrder.CloseOrder()
	}

	if sellingOrder.PendingShares == 0 {
		sellingOrder.CloseOrder()
	}

	b.Transactions = append(b.Transactions, transaction)
}
