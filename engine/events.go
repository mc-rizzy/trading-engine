package engine

import "time"

// EventType categorizes what happened in the engine
type EventType string

const (
    EventOrderFilled   EventType = "ORDER_FILLED"
    EventOrderPartial  EventType = "ORDER_PARTIAL"
    EventOrderCanceled EventType = "ORDER_CANCELED"
    EventTradeExecuted EventType = "TRADE_EXECUTED"
)

// OrderEvent payload for state tracking changes
type OrderEvent struct {
    Type           EventType
    OrderID        string
    UserID         string
    FilledQuantity uint32
    Status         string
    Timestamp      time.Time
}

// TradeEvent payload generated when two orders successfully match
type TradeEvent struct {
    Type         EventType
    TradeID      string
    MakerOrderID string
    TakerOrderID string
    BuyerID      string
    SellerID     string
    AssetType    string
    Price        uint64
    Quantity     uint32
    Timestamp    time.Time
}

// EventDispatcher manages the asynchronous event pipelines
type EventDispatcher struct {
    OrderChannel chan OrderEvent
    TradeChannel chan TradeEvent
}

// NewEventDispatcher initializes buffered channels to handle traffic spikes safely
func NewEventDispatcher(bufferSize int) *EventDispatcher {
    return &EventDispatcher{
        OrderChannel: make(chan OrderEvent, bufferSize),
        TradeChannel: make(chan TradeEvent, bufferSize),
    }
}