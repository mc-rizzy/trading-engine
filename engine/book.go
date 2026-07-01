package engine

import "time"

// Side determines if the order is buying or selling
type Side string
const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

// Order represents an individual customer order ticket
type Order struct {
	ID        string
	UserID    string
	Side      Side
	Price     uint64    // Stored in cents (e.g., $10.50 is 1050)
	Quantity  uint32
	Timestamp time.Time
	Next      *Order    // Pointer to the order behind this one in line
	Prev      *Order    // Pointer to the order ahead of this one in line
}

// Limit represents a single price level containing a queue of orders
type Limit struct {
	Price     uint64
	HeadOrder *Order    // The front of the line (oldest order)
	TailOrder *Order    // The back of the line (newest order)
}

// OrderBook manages all active limits and offers O(1) order lookups
type OrderBook struct {
	Bids   map[uint64]*Limit // Maps a Buy Price -> The Limit line
	Asks   map[uint64]*Limit // Maps a Sell Price -> The Limit line
	Orders map[string]*Order // Maps an OrderID -> The Order node for instant cancellations
}

// NewOrderBook initializes an empty order book
func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids:   make(map[uint64]*Limit),
		Asks:   make(map[uint64]*Limit),
		Orders: make(map[string]*Order),
	}
}