/*
  btcrobot is a Bitcoin, Litecoin and Altcoin trading bot written in golang,
  it features multiple trading methods using technical analysis.

  Disclaimer:

  USE AT YOUR OWN RISK!

  The author of this project is NOT responsible for any damage or loss caused
  by this software. There can be bugs and the bot may not Tick as expected
  or specified. Please consider testing it first with paper trading /
  backtesting on historical data. Also look at the code to see what how
  it's working.

  Weibo:http://weibo.com/bocaicfa
*/

package strategy

import (
	. "common"
	. "config"
	"email"
	"fmt"
	"logger"
	"strconv"
)

type the3crowStrategy struct {
	PrevClosePrice float64
}

func init() {
	the3crow := new(the3crowStrategy)
	Register("the3crow", the3crow)
}

//the3crow strategy
func (the3crow *the3crowStrategy) Tick(records []Record) bool {
	//read config

	tradeAmount := Option["tradeAmount"]

	numTradeAmount, err := strconv.ParseFloat(Option["tradeAmount"], 64)
	if err != nil {
		logger.Errorln("config item tradeAmount is not float")
		return false
	}

	var Time []string
	var Price []float64
	var Volumn []float64
	for _, v := range records {
		Time = append(Time, v.TimeStr)
		Price = append(Price, v.Close)
		Volumn = append(Volumn, v.Volumn)
	}

	length := len(Price)
	if the3crow.PrevClosePrice == records[length-1].Close {
		return false
	}

	the3crow.PrevClosePrice = records[length-1].Close

	logger.Infof("nowClose %0.02f\n", records[length-1].Close)
	logger.Infof("3 open %0.02f close %0.02f\n", records[length-2].Open, records[length-2].Close)
	logger.Infof("2 open %0.02f close %0.02f\n", records[length-3].Open, records[length-3].Close)
	logger.Infof("1 open %0.02f close %0.02f\n", records[length-4].Open, records[length-4].Close)

	if records[length-2].Close > records[length-2].Open {
		logger.Infof("3阳")
	} else {
		logger.Infof("3阴")
	}

	if records[length-3].Close > records[length-3].Open {
		logger.Infof("2阳")
	} else {
		logger.Infof("2阴")
	}
	if records[length-4].Close > records[length-4].Open {
		logger.Infof("1阳")
	} else {
		logger.Infof("1阴")
	}
	logger.Infoln("---------")

	//the3crow cross
	if records[length-2].Close > records[length-2].Open &&
		records[length-3].Close > records[length-3].Open &&
		records[length-4].Close > records[length-4].Open {
		if Option["enable_trading"] == "1" && PrevTrade != "buy" {
			if GetAvailable_cny() < numTradeAmount {
				warning = "the3crow up, but 没有足够的法币可买"
				PrevTrade = "buy"
			} else {
				var tradePrice string
				if true {
					ret, orderBook := GetOrderBook()
					if !ret {
						logger.Infoln("get orderBook failed 1")
						ret, orderBook = GetOrderBook() //try again
						if !ret {
							logger.Infoln("get orderBook failed 2")
							return false
						}
					}

					logger.Infoln("卖一", (orderBook.Asks[len(orderBook.Asks)-1]))
					logger.Infoln("买一", orderBook.Bids[0])

					tradePrice = fmt.Sprintf("%f", orderBook.Bids[0].Price+0.01)
					warning += "---->限价单" + tradePrice
				} else {
					tradePrice = getTradePrice("buy", Price[length-1])
					warning += "---->市价单" + tradePrice
				}

				buyID := Buy(tradePrice, tradeAmount)
				if buyID != "0" {
					warning += "[委托成功]" + buyID
				} else {
					warning += "[委托失败]"
				}
			}

			logger.Infoln(warning)
			go email.TriggerTrender(warning)
		}
	} else if records[length-2].Close < records[length-2].Open &&
		records[length-3].Close < records[length-3].Open &&
		records[length-4].Close < records[length-4].Open {
		if Option["enable_trading"] == "1" && PrevTrade != "sell" {
			if GetAvailable_coin() < numTradeAmount {
				warning = "the3crow down, but 没有足够的币可卖"
				PrevTrade = "sell"
				PrevBuyPirce = 0
			} else {
				warning = "the3crow down, 卖出Sell Out---->市价" + getTradePrice("", Price[length-1])
				var tradePrice string
				if true {
					ret, orderBook := GetOrderBook()
					if !ret {
						logger.Infoln("get orderBook failed 1")
						ret, orderBook = GetOrderBook() //try again
						if !ret {
							logger.Infoln("get orderBook failed 2")
							return false
						}
					}

					logger.Infoln("卖一", (orderBook.Asks[len(orderBook.Asks)-1]))
					logger.Infoln("买一", orderBook.Bids[0])

					tradePrice = fmt.Sprintf("%f", orderBook.Asks[len(orderBook.Asks)-1].Price-0.01)
					warning += "---->限价单" + tradePrice
				} else {
					tradePrice = getTradePrice("sell", Price[length-1])
					warning += "---->市价单" + tradePrice
				}

				sellID := Sell(tradePrice, tradeAmount)
				if sellID != "0" {
					warning += "[委托成功]"
				} else {
					warning += "[委托失败]"
				}
			}

			logger.Infoln(warning)
			go email.TriggerTrender(warning)
		}
	}

	//do sell when price is below stoploss point
	processStoploss(Price)

	processTimeout()

	return true
}