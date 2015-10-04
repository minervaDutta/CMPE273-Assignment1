//Pgm to 
//Request for new stocks given the percentage of stocks to buy from a budget
//View portfolio
//Trade ID is machine dependednt (Mac)

package main

import (
	"os"
	"strings"
	"fmt"
	"strconv"
	"net"
	"net/rpc/jsonrpc"
)

type RequestMap struct{
	Stocks map[string]float64
}

type ResponseList struct{
	TradeID string
	NoOfStocks []int
	Price []float64
	Symbol []string
	Unvested []float64
}

type RequestTradeID struct{
 	TradeId string
 }

type ResponseTradeID struct{
	Symbol []string
	CurrentPrice []float64
	ChangeInPrice []float64
	Unvested []float64
	NoOfStocks []int
}



func main() {
	stockMap := make(map[string]float64)
	conn, err := net.Dial("tcp", "localhost:8222")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	cn := jsonrpc.NewClient(conn)

// commandline arguments passed through the client as 
//go run yahooFinanceClient.go GOOG:25,IBM:25,VMW:50 10000

if len(os.Args)>2{

	strOne := strings.Split(os.Args[1], ",")
	strTwo := os.Args[2]
	budget, err := strconv.Atoi(strTwo)
	if (err!=nil){
		panic("Fatal Error!")
	}
	for i:=0; i< len(strOne); i++{
	
		eachSym := strings.Split(strOne[i],":")		
		percent:= strings.TrimSuffix(eachSym[1],"%")
		percentage, err1 := strconv.Atoi(percent)
		
		if (err1!=nil){
			panic("Fatal Error!")
		}
		allotedMoney := float64(budget * percentage/100)
		stockMap[eachSym[0]] = allotedMoney
	}

	var request *RequestMap
	var response ResponseList
	request = &RequestMap{stockMap}
    err = cn.Call("Finance.DoTheJob", request, &response)
	if err != nil {
		fmt.Errorf("finance error:", err)
	}

	fmt.Println("TradeID: ", response.TradeID)
	var finalString string
	var totalUnvested float64
	for i:=range response.Symbol{

		finalString += fmt.Sprintf("%s:%d:$%g, ", response.Symbol[i],response.NoOfStocks[i],response.Price[i])
		totalUnvested +=response.Unvested[i]
	}
	fmt.Println("stocks: ", finalString)
	fmt.Println("Total unvested amount: ", totalUnvested)


}else{
	// in case of requesting for portfolio

	getTradeID:=os.Args[1]
	var requestId *RequestTradeID
	requestId = &RequestTradeID{getTradeID}
	var responseId ResponseTradeID
	err = cn.Call("Finance.GetPortfolio", requestId, &responseId)
	if err != nil {
		fmt.Errorf("finance error:", err)
	}
	var finalPortfolioString, finalCurrentValue string
	var totalPortfolioUnvested float64
	for i:=range responseId.Symbol{
		if responseId.ChangeInPrice[i]>0{

			finalPortfolioString += fmt.Sprintf("%s:%d:+$%g, ", responseId.Symbol[i],responseId.NoOfStocks[i],responseId.CurrentPrice[i])
		}else{
			finalPortfolioString += fmt.Sprintf("%s:%d:-$%g, ", responseId.Symbol[i],responseId.NoOfStocks[i],responseId.CurrentPrice[i])
		}

		finalCurrentValue += fmt.Sprintf("%s:$%g, ",responseId.Symbol[i], responseId.CurrentPrice[i])
		totalPortfolioUnvested +=responseId.Unvested[i]
	}
	fmt.Println(finalPortfolioString)
	fmt.Println("CurrentMarketPrice: ", finalCurrentValue)
	fmt.Println("Total unvested amount:", totalPortfolioUnvested)


}
	
}
