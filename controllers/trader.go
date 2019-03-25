package controllers

import (
	"github.com/astaxie/beego"
	"encoding/json"
	"github.com/xfrzrcj/huobi_trader/models"
	"crypto/md5"
	"fmt"
	"strings"
	"github.com/astaxie/beego/logs"
	"github.com/xfrzrcj/huobi_trader/constants"
	"github.com/xfrzrcj/huobi_trader/untils"
)

type TraderController struct {
	beego.Controller
}

const (
	salt = ""
	IP_WHITE_LIST = "ip_white_list"
	SPLITE_STR = ","

)
var (
	logger = logs.GetBeeLogger()
	ipList = getIpList()
	Key = []byte{}
	iv = ""
)
func (o *TraderController) Post() {
	logger.Info("request:%s",string(o.Ctx.Input.RequestBody))
	var data map[string]interface{}
	addr := o.Ctx.Request.RemoteAddr
	logger.Debug(addr)
	if checkAddr(addr){
		if request,err := untils.AesDecryptSimple(string(o.Ctx.Input.RequestBody),Key,iv);err==nil{
			logger.Debug("decodeRequest:"+string(request))
			var ob models.Trader
			json.Unmarshal(request, &ob)
			if ob.Amount <= constants.ZERO {
				data = getData(nil, constants.PARAM_ERROR)
			} else {
				ob.Coin = strings.ToLower(ob.Coin)
				res, err := models.ProcessTrade(&ob)
				logger.Debug(err.Message)
				data = getData(res, err)
			}
		}else{
			logger.Error("%s",err.Error())
			data = getData(nil, constants.PARAM_ERROR)
		}
	}else{
		logger.Debug("ip error")
		data = getData(nil, constants.PARAM_ERROR)
	}
	if json,jsonErr := json.Marshal(&data);jsonErr==nil{
		logger.Info("response:%s",string(json))
		if res,err := untils.AesEncryptSimple(json,Key,iv);err==nil{
			o.Ctx.Output.Body([]byte(res))
		}
		o.Ctx.Output.Body(json)
	}else{
		o.Ctx.Output.Body([]byte{})
	}
}

func getData(detail interface{}, err constants.ApiError) (data map[string]interface{}) {
	if err.Code == constants.SUCCESS.Code {
		data = map[string]interface{}{"result_code": err.Code, "message": err.Message, "data": detail}
	} else {
		data = map[string]interface{}{"result_code": err.Code, "message": err.Message}
	}
	return data
}

func md5str(msg string) (string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(msg)))
}

func checkSign(ts string,uid string,body string,sign string)bool{
	return true
	bodyMd5 := md5str(body)
	tsuidMd5 := md5str(ts + uid + salt)
	calSig := md5str(bodyMd5 + tsuidMd5)
	logger.Debug("ts:"+ts)
	logger.Debug("sign:"+sign)
	logger.Debug("uid:"+uid)
	logger.Debug("calsign:"+calSig)
	return strings.Compare(calSig,sign)==0
}

func getIpList() []string{
	whiteList := beego.AppConfig.String(IP_WHITE_LIST)
	return strings.Split(whiteList,SPLITE_STR)
}

func checkAddr(addr string)bool{
	if strings.Compare("*",ipList[0])==0 {
		return true
	}
	ip := strings.Split(addr,":")[0]
	length := len(ipList)
	for i:=0;i< length;i++{
		if strings.Compare(ipList[i],ip)==0{
			return true
		}
	}
	return false
}
