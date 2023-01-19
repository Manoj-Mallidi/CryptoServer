package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

type Currency struct {
	ID          string `json:"id,omitempty"`
	FullName    string `json:"fullName,omitempty"`
	Ask         string `json:"ask"`
	Bid         string `json:"bid"`
	Last        string `json:"last"`
	Open        string `json:"open"`
	Low         string `json:"low"`
	High        string `json:"high"`
	FeeCurrency string `json:"feeCurrency"`
}

func currencyHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Path[len("/currency/"):]
	switch r.Method {
	case "GET":
		if symbol == "all" {
			getAllCurrencies(w, r)
		} else {
			getCurrency(w, r, symbol)
		}
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
func getAllCurrencies(w http.ResponseWriter, r *http.Request) {
	// Use the hitbtc API to get all currency information
	resp, err := http.Get("https://api.hitbtc.com/api/2/public/ticker")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body and unmarshal it into a struct
	var currencies []Currency
	if err := json.NewDecoder(resp.Body).Decode(&currencies); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the data into JSON and write it to the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currencies)
}

func getCurrency(w http.ResponseWriter, r *http.Request, symbol string) {
	// Use the hitbtc API to get information for the specific symbol
	url := fmt.Sprintf("https://api.hitbtc.com/api/2/public/ticker/%s", symbol)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body and unmarshal it into a struct
	var currency Currency
	if err := json.NewDecoder(resp.Body).Decode(&currency); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the data into JSON and write it to the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currency)
}
func main() {
	http.HandleFunc("/currency/", currencyHandler)
	http.ListenAndServe(":8080", nil)

	ws, err := websocket.Dial("wss://api.hitbtc.com/api/2/ws", "", "")
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws.Close()

	// Subscribe to the market data feed
	msg := []byte(`{"method":"subscribeTicker","params":["BTCUSD"],"id":1}`)
	if _, err := ws.Write(msg); err != nil {
		log.Println("write:", err)
	}

	// Create a map to store the currency data in memory
	currencies := make(map[string]Currency)

	// Continuously read messages from the websocket
	for {
		message := make([]byte, 2048)
		n, err := ws.Read(message)
		if err != nil {
			log.Println("read:", err)
			break
		}
		message = message[:n]

		// Unmarshal the message into a struct
		var tickerData struct {
			Method string `json:"method"`
			Params struct {
				Symbol string `json:"symbol"`
			} `json:"params"`
			Result Currency `json:"result"`
		}
		if err := json.Unmarshal(message, &tickerData); err != nil {
			log.Println("unmarshal:", err)
			continue
		}

		// Update the currency data in the map
		currencies[tickerData.Params.Symbol] = tickerData.Result
	}
}
