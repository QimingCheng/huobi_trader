package constants

type ApiError struct {
	Code int8
	Message string
}

var NOT_ENOUGH_DEPTH = ApiError{2,"交易深度不足，无法出售"}
var SUCCESS = ApiError{1,"success"}
var PARAM_ERROR = ApiError{3,"参数错误"}
var SYMBOL_ERROR = ApiError{4,"交易对不存在"}
var DB_ERROR = ApiError{5,"数据库错误"}
var NO_SUCH_ACCOUNT = ApiError{6,"对应币账户不存在"}
var SYS_ERROR = ApiError{0,"系统错误"}
var BALANCE_NOT_ENOUGH = ApiError{7,"账户余额不足"}

const (
	EMPTY_STR = ""
	ZERO = 0
)
