package models

import(
	"github.com/xfrzrcj/huobi_trader/service"
	"github.com/xfrzrcj/huobi_trader/constants"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/satori/go.uuid"
	"time"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
	"github.com/xfrzrcj/huobi_trader/models/huobi"
	"fmt"
)
const (
	BASE = 10
	TRADE_MOENY = "trade"
	BIT_SIZE_64 = 64
	USDT = "usdt"
	SELL_MARKET = "sell-market"
	API_STR = "api"
	QUERY_TIMES = 5
	OK_STR = "ok"
)

const(
	ORDER_INIT = 0
	ORDER_START = 1
	ORDER_FILLED = 2
	ORDER_FIELD = 3
)
func init(){
	//orm.RegisterDataBase("default","mysql","root:Dd112211!@tcp(116.62.236.199:3306)/huobi_trade?charset=utf8",10,10)
	orm.RegisterModel(new(TradeDetail))
	getSymbol()
	//orm.RunSyncdb("default", false, true)
}

var (
	Traders map[string]*Trader
	logger = logs.GetBeeLogger()
	symbolMap = make(map[string]huobi.SymbolsData)
)

type Trader struct {
	Coin	string	`json:"coin"`
	Amount	float64	`json:"amount"`
}

type TradeDetail struct{
	Id     	string `orm:"pk" json:"id"`
	Coin	string `json:"coin"`
	Amount	float64 `orm:"digits(12);decimals(4)" json:"amount"`
	Price	float64 `orm:"digits(18);decimals(10)" json:"price"`
	Sum 	float64 `orm:"digits(20);decimals(10)" json:"sum"`
	Ts		int64 `json:"ts"`
	Status 	int8 `json:"status"`
	RealPrice	string `json:"real_price"`
	RealAmount 	string `json:"real_amount"`
	RealSum 	string `json:"real_sum"`
	Fee 		string `json:"fee"`
	OrderId		string `json:"order_id"`
}

func ProcessTrade(trade *Trader) (detail TradeDetail,err constants.ApiError){
	//detail.Id = "7335cd0c-0463-4441-b64b-17074ab2d621"
	//detail.Coin = trade.Coin
	//detail.Amount = trade.Amount
	//detail.Ts = time.Now().Unix()
	//detail.Sum = float64(1232.345)
	//detail.Price = float64(23.45)
	//return detail,constants.SUCCESS
	//获取账户余额
	balance,accountId,err := getBalance(trade.Coin)
	if err!=constants.SUCCESS{
		return detail, err
	}
	logger.Info("balance:%f,account_id:%d",balance,accountId)
	//检查数量精度是否符合要求
	symbol := trade.Coin + USDT
	fmtStr := "%."+strconv.Itoa(symbolMap[symbol].AmountPrecision)+"f"
	logger.Debug("fmtStr:%s",fmtStr)
	amount := fmt.Sprintf(fmtStr,trade.Amount)
	i, errs := strconv.ParseFloat(amount, BIT_SIZE_64)
	logger.Debug("formatAmount:%f",i)
	logger.Debug("detail.Amount:%f",trade.Amount)
	if errs !=nil || i != trade.Amount{
		return detail, constants.PARAM_ERROR
	}
	//获取市场深度
	depth := service.GetMarketDepth(trade.Coin + USDT,"step0")
	bids := depth.Tick.Bids
	if len(bids)==0{
		return detail, constants.SYMBOL_ERROR
	}
	var count = float64(0)
	var temp = float64(0)
	var sum = float64(0)
	//预估总价
	for _,value := range bids{
		temp = count + value[1]
		if temp >= trade.Amount {
			sum = sum + value[0] * (trade.Amount - count)
			count = trade.Amount
			break
		}
		sum = sum + value[0] * value[1]
		count = temp
	}
	//检查余额是否充足
	if sum > balance {
		return detail, constants.BALANCE_NOT_ENOUGH
	}
	if count == trade.Amount {
		detail.Price = sum/trade.Amount
		detail.Sum = sum
		detail.Coin = trade.Coin
		detail.Amount = trade.Amount
		detail.Id = uuid.Must(uuid.NewV4(),nil).String()
		detail.Ts = time.Now().Unix()
		detail.Status = ORDER_INIT
		err = dealOrder(&detail,accountId)
		return detail,err
	}else{
		return detail, constants.NOT_ENOUGH_DEPTH
	}
}

func dealOrder(detail *TradeDetail,accountId int64) constants.ApiError{
	dbOrm := orm.NewOrm()
	dbOrm.Begin()
	if _,err := dbOrm.Insert(detail);err == nil{
		symbol := detail.Coin+USDT
		fmtStr := "%."+strconv.Itoa(symbolMap[symbol].AmountPrecision)+"f"
		amount := fmt.Sprintf(fmtStr,float64(detail.Amount))
		param := huobi.PlaceRequestParams{Amount:amount,Type:SELL_MARKET,Symbol:symbol,AccountID:strconv.FormatInt(accountId,BASE),Source:API_STR}
		placeReturn := service.Place(param)
		if strings.Compare(placeReturn.Status,OK_STR)==0{
			detail.OrderId = placeReturn.Data
			detail.Status = ORDER_START
			if _,err := dbOrm.Update(detail);err != nil{
				dbOrm.Rollback()
				logger.Error("update order to start error!id:%s,orderId:%s",detail.Id,detail.OrderId)
			}else{
				dbOrm.Commit()
			}
			//更新订单成交情况
			go QueryOrderStatus(detail,dbOrm)
			return constants.SUCCESS
		}else{
			dbOrm.Rollback()
			return constants.ApiError{Code:-1,Message:placeReturn.ErrMsg}
		}
	}else{
		dbOrm.Rollback()
		return constants.DB_ERROR
	}
}

func QueryOrderStatus(detail *TradeDetail,dbOrm orm.Ormer){
	notFilled := true
	for i := 0;i< QUERY_TIMES;i++{
		orderReturn := service.QueryOrder(detail.OrderId)
		if strings.Compare("filled",orderReturn.Data.State)==0 {
			detail.Status = ORDER_FILLED
			detail.RealAmount = orderReturn.Data.FieldAmount
			detail.RealSum = orderReturn.Data.FieldCashAmount
			detail.Fee = orderReturn.Data.FieldFee
			if sum,sumErr := strconv.ParseFloat(detail.RealSum,BIT_SIZE_64);sumErr==nil{
				if amount,amountErr := strconv.ParseFloat(detail.RealAmount,BIT_SIZE_64);amountErr==nil{
					symbol := detail.Coin+USDT
					fmtStr := "%."+strconv.Itoa(symbolMap[symbol].PricePrecision)+"f"
					detail.RealPrice = fmt.Sprintf(fmtStr,sum/amount)
				}
			}
			notFilled = false
			break
		}
		time.Sleep(500)
	}
	if notFilled {
		detail.Status = ORDER_FIELD
		logger.Error("order not filled,orderId:!" + detail.OrderId)
	}
	if _,err := dbOrm.Update(detail);err != nil{
		logger.Error("update order to filled error!",err)
	}
	logger.Debug("end dealOrder,time:%d",time.Now().Unix())
}

func getSymbol(){
	symbolList := service.GetSymbols().Data
	for i:=0;i< len(symbolList);i++{
		symbolMap[symbolList[i].Symbol] = symbolList[i]
	}
}

func getBalance(coin string) (balance float64,accountId int64,err constants.ApiError){
	accounts := service.GetAccounts()
	if len(accounts.Data) > 0 {
		balanceReturn := service.GetAccountBalance(strconv.FormatInt(accounts.Data[0].ID,BASE)).Data
		list := balanceReturn.List
		accountId = balanceReturn.ID
		for i:=0;i< len(list);i++{
			if strings.Compare(list[i].Currency,coin)==0 && strings.Compare(list[i].Type,TRADE_MOENY)==0{
				var err error
				logger.Debug("currency:%s,type:%s,balance:%s",list[i].Currency,list[i].Type,list[i].Balance)
				if balance,err = strconv.ParseFloat(list[i].Balance, BIT_SIZE_64);err==nil{
					return balance,accountId, constants.SUCCESS
				}else{
					return balance,accountId, constants.SYS_ERROR
				}
			}
		}
		return balance,accountId, constants.BALANCE_NOT_ENOUGH
	}else{
		return balance,accountId, constants.NO_SUCH_ACCOUNT
	}
}
