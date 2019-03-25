package huobi

type OrderDetail struct{
	Id 			int64  	`json:"id"`
	Symbol 		string 	`json:symbol`
	AccountId	int64  	`json:"account-id"`
	Amount 		string 	`json:"amount"`
	Price 		string 	`json:"price"`
	CreateAt	int64  	`json:"created-at"`
	Type 		string 	`json:"type"`
	FieldAmount	string 	`json:"field-amount"`
	FieldCashAmount	string	`json:"field-cash-amount"`
	FieldFee	string 	`json:"field-fees"`
	FinishedAt	int64	`json:"finished-at"`
	Source 		string 	`json:"source"`
	State 		string 	`json:"state"`
	CancelAt 	int64 	`json:"canceled-at"`
}

type OrderDetailReturn struct {
	Status  string `json:"status"`
	Data    OrderDetail `json:"data"`
	ErrCode string `json:"err-code"`
	ErrMsg  string `json:"err-msg"`
}