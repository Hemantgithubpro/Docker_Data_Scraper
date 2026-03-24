package main

import (
	// "fmt"
	// "context"
	"log"
	// "time"
)

func main() {
	apikey, jwtToken, clientId, feedToken, err := getCredentials()
	if err != nil {
		log.Printf("Error getting credentials: %v", err)
	}
	if jwtToken == "" || apikey == "" || clientId == "" || feedToken == "" {
		log.Fatal("Missing required environment variables: jwt_token, API_KEY, CLIENT_ID, feed_token")
	}


	// // Start WebSocket Connection
	tokens := []TokenInfo{
		// {ExchangeType: 1, Tokens: []string{"99926000","2885"}}, // nifty 50, reliance nse
		// {ExchangeType: 3, Tokens: []string{"99919000"}}, // sensex bse
		{ExchangeType: 2, Tokens: []string{"62434"}}, // nfo 
		// {ExchangeType: 1, Tokens: []string{"2885"}},
	}

	websocketprint(jwtToken, apikey, clientId, feedToken, 1, tokens)
}
