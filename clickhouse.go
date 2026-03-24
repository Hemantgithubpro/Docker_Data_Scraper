package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouseWriter struct {
	conn driver.Conn
}

func NewClickHouseWriterFromEnv() (*ClickHouseWriter, error) {
	host := getenvDefault("CLICKHOUSE_HOST", "clickhouse")
	port := getenvDefault("CLICKHOUSE_PORT", "9000")
	database := getenvDefault("CLICKHOUSE_DB", "market_data")
	username := getenvDefault("CLICKHOUSE_USER", "market_user")
	password := getenvDefault("CLICKHOUSE_PASSWORD", "market_password")

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", host, port)},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		DialTimeout:     5 * time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	writer := &ClickHouseWriter{conn: conn}
	if err := writer.ensureSchema(ctx); err != nil {
		return nil, err
	}

	return writer, nil
}

func (w *ClickHouseWriter) Close() {
	if w == nil || w.conn == nil {
		return
	}
	_ = w.conn.Close()
}

func (w *ClickHouseWriter) ensureSchema(ctx context.Context) error {
	return w.conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS market_ticks (
			token String,
			exchange_type Int32,
			timestamp DateTime64(3, 'UTC'),
			ltp Float64,
			volume Int64,
			open_price Float64,
			high_price Float64,
			low_price Float64,
			close_price Float64,
			total_buy_qty Float64,
			total_sell_qty Float64,
			avg_traded_price Float64,
			upper_circuit Float64,
			lower_circuit Float64,
			high_52_week Float64,
			low_52_week Float64
		)
		ENGINE = MergeTree
		ORDER BY (token, timestamp)
	`)
}

func (w *ClickHouseWriter) InsertTicks(ctx context.Context, ticks []MarketTick) error {
	if len(ticks) == 0 {
		return nil
	}

	batch, err := w.conn.PrepareBatch(ctx, `
		INSERT INTO market_ticks (
			token, exchange_type, timestamp, ltp, volume,
			open_price, high_price, low_price, close_price,
			total_buy_qty, total_sell_qty, avg_traded_price,
			upper_circuit, lower_circuit, high_52_week, low_52_week
		)
	`)
	if err != nil {
		return err
	}

	for _, tick := range ticks {
		err := batch.Append(
			tick.Token,
			int32(tick.ExchangeType),
			tick.Timestamp.UTC(),
			tick.LTP,
			tick.Volume,
			tick.OpenPrice,
			tick.HighPrice,
			tick.LowPrice,
			tick.ClosePrice,
			tick.TotalBuyQty,
			tick.TotalSellQty,
			tick.AvgTradedPrice,
			tick.UpperCircuit,
			tick.LowerCircuit,
			tick.High52Week,
			tick.Low52Week,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

func flushBufferPeriodically(buffer *TickBuffer, writer *ClickHouseWriter, interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ticks := buffer.Flush()
			if len(ticks) == 0 {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := writer.InsertTicks(ctx, ticks)
			cancel()
			if err != nil {
				log.Printf("clickhouse insert failed for batch of %d ticks: %v", len(ticks), err)
			}
		case <-stop:
			return
		}
	}
}

func getenvDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
