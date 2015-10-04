package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"os"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"encoding/json"
	
)

const (
	timeout = time.Duration(time.Second * 10)
)

type RequestMap struct{
	Stocks map[string]float64
}

type RequestTradeID struct{
 	TradeId string
 }

type Stock struct {
	List struct {
		Resources []struct {
			Resource struct {
				Fields struct {
					Name    string `json:"name"`
					Price   string `json:"price"`
					Symbol  string `json:"symbol"`
					Ts      string `json:"ts"`
					Type    string `json:"type"`
					UTCTime string `json:"utctime"`
					Volume  string `json:"volume"`
				} `json:"fields"`
			} `json:"resource"`
		} `json:"resources"`
	} `json:"list"`
}

type ResponseList struct{
	TradeID string
	NoOfStocks []int
	Price []float64
	Symbol []string
	Unvested []float64
}

type ResponseTradeID struct{
	Symbol []string
	CurrentPrice []float64
	ChangeInPrice []float64
	Unvested []float64
	NoOfStocks []int
}

type Data struct{
				Price float64
				UnvestedAmount float64
				NumberOfStocks int
			}
var data Data
var tradeMap map[string]map[string] Data
var map1 map[string]Data
var tempMap map[string]Data


type Finance int

func (f *Finance) DoTheJob(request *RequestMap, response *ResponseList ) error {

 	///For Generating a unique UUID
	tradeID, err := newUUID()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	response.TradeID = string(tradeID)
	tradeMap= make(map[string]map[string] Data)	
	tempMap=make(map[string]Data)
	


	//Fetching the count of stocks and price
	for key, value:= range request.Stocks{
		stockPrice := getQuote(key)
		count:= int(value/stockPrice)
		unvestedAmont := value - float64(count)*stockPrice

		response.Symbol = append(response.Symbol, key)
		response.Price = append(response.Price, stockPrice)
		response.NoOfStocks = append(response.NoOfStocks, int(count))
		response.Unvested = append(response.Unvested, unvestedAmont)

		data = Data{stockPrice,unvestedAmont,int(count)}
		tempMap[key]= data
	}
	tradeMap[tradeID]=tempMap
	return nil
}

// dependent on mac OSX

func newUUID() (string, error) {
	f, _ := os.Open("/dev/urandom")
	b := make([]byte, 16)
	f.Read(b)
	f.Close()
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil

}


func getQuote(symbol string) float64 {
	// set http timeout
	client := http.Client{Timeout: timeout}

	url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json", symbol)
	res, err := client.Get(url)
	if err != nil {
		fmt.Errorf("Stocks cannot access the yahoo finance API: %v", err)
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Errorf("Stocks cannot read json body: %v", err)
	}

	var stock Stock

	err = json.Unmarshal(content, &stock)
	if err != nil {
		fmt.Errorf("Stocks cannot parse the json data: %v", err)
	}

	price, err := strconv.ParseFloat(stock.List.Resources[0].Resource.Fields.Price, 64)
	if err != nil {
		fmt.Errorf("Stock price: %v", err)
	}

	return price
}



func (f *Finance) GetPortfolio(requestId *RequestTradeID, responseId *ResponseTradeID ) error {
	map1 := tradeMap[requestId.TradeId]
	for key,value:= range map1{

		updatedPrice := getQuote(key)
		previousPrice := value.Price
		changeInPrice:=updatedPrice - previousPrice		
		responseId.Symbol = append(responseId.Symbol, key)
		responseId.CurrentPrice = append(responseId.CurrentPrice, updatedPrice)
		responseId.ChangeInPrice = append(responseId.ChangeInPrice, changeInPrice)
		responseId.Unvested = append(responseId.Unvested, value.UnvestedAmount)
		responseId.NoOfStocks = append(responseId.NoOfStocks, value.NumberOfStocks)
	}
	return nil

}


func main() {

	fin := new(Finance)
	server := rpc.NewServer()
	server.Register(fin)

	l, e := net.Listen("tcp", ":8222")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}