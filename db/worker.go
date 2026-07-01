package db

import (
    "context"
    "database/sql"
    "log"
    "trading-engine/engine"
)

type DatabaseWorker struct {
    db         *sql.DB
    dispatcher *engine.EventDispatcher
}

func NewDatabaseWorker(db *sql.DB, dispatcher *engine.EventDispatcher) *DatabaseWorker {
    return &DatabaseWorker{db: db, dispatcher: dispatcher}
}

// Start spawns background routines listening for matching engine outputs
func (w *DatabaseWorker) Start(ctx context.Context) {
    // Worker handling trade persistence
    go func() {
        for {
            select {
            case trade := <-w.dispatcher.TradeChannel:
                err := w.persistTrade(ctx, trade)
                if err != nil {
                    log.Printf("CRITICAL: Failed to persist trade %s: %v", trade.TradeID, err)
                    // In real fintech, route failed transactions to a Dead-Letter Queue (DLQ) for manual audit
                }
            case <-ctx.Done():
                return
            }
        }
    }()
}

// persistTrade wraps our SQL statements into an atomic transaction
func (w *DatabaseWorker) persistTrade(ctx context.Context, t engine.TradeEvent) error {
    tx, err := w.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback() // Rollback is skipped if tx.Commit() is reached successfully

    // 1. Insert the trade record
    tradeQuery := `INSERT INTO trades (id, maker_order_id, taker_order_id, buyer_id, seller_id, asset_type, price, quantity, executed_at) 
                   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
    _, err = tx.ExecContext(ctx, tradeQuery, t.TradeID, t.MakerOrderID, t.TakerOrderID, t.BuyerID, t.SellerID, t.AssetType, t.Price, t.Quantity, t.Timestamp)
    if err != nil {
        return err
    }

    // 2. Adjust asset balances atomically
    // Deduct locked fiat currency from buyer
    _, err = tx.ExecContext(ctx, "UPDATE accounts SET locked_balance = locked_balance - $1 WHERE user_id = $2 AND asset_type = 'USD'", t.Price*uint64(t.Quantity), t.BuyerID)
    if err != nil {
        return err
    }
    
    // Add traded asset to buyer
    _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE user_id = $2 AND asset_type = $3", t.Quantity, t.BuyerID, t.AssetType)
    if err != nil {
        return err
    }

    // Deduct locked asset from seller
    _, err = tx.ExecContext(ctx, "UPDATE accounts SET locked_balance = locked_balance - $1 WHERE user_id = $2 AND asset_type = $3", t.Quantity, t.SellerID, t.AssetType)
    if err != nil {
        return err
    }

    // Add fiat currency payment to seller
    _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE user_id = $2 AND asset_type = 'USD'", t.Price*uint64(t.Quantity), t.SellerID)
    if err != nil {
        return err
    }

    return tx.Commit()
}