package main

import (
	"context"
	"log"
	"time"
)

func main() {
	apikey, jwtToken, clientId, feedToken, err := getCredentials()
	if err != nil {
		log.Printf("Error getting credentials: %v", err)
	}
	if jwtToken == "" || apikey == "" || clientId == "" || feedToken == "" {
		log.Fatal("Missing required environment variables: jwt_token, API_KEY, CLIENT_ID, feed_token")
	}

	writer, err := NewClickHouseWriterFromEnv()
	if err != nil {
		log.Fatalf("failed to initialize clickhouse writer: %v", err)
	}
	defer writer.Close()

	buffer := NewTickBuffer()
	stopFlush := make(chan struct{})
	flushDone := make(chan struct{})

	go func() {
		defer close(flushDone)
		flushBufferPeriodically(buffer, writer, 2*time.Second, stopFlush)
	}()

	// // Start WebSocket Connection
	tokens := []TokenInfo{
		// {ExchangeType: 1, Tokens: []string{"99926000","2885"}}, // nifty 50, reliance nse
		// {ExchangeType: 3, Tokens: []string{"99919000"}}, // sensex bse
		{ExchangeType: 2, Tokens: []string{"62434"}}, // nfo
		// {ExchangeType: 1, Tokens: []string{"2885"}},
	}

	websocketConnectiontoDB(jwtToken, apikey, clientId, feedToken, 1, tokens, buffer)

	close(stopFlush)
	<-flushDone

	remaining := buffer.Flush()
	if len(remaining) > 0 {
		if err := writer.InsertTicks(context.Background(), remaining); err != nil {
			log.Printf("final flush failed: %v", err)
		}
	}
}
