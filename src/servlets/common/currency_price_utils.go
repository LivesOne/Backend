package common

import (
	"github.com/garyburd/redigo/redis"
	"utils"
	"utils/logger"
)

const(
	CURRENCY_PRICE_RDS_KEY = "cache:currencyprice"
	CURRENCY_PRICE_RDS_EXPIRE = 60*30
)


type CurrencyPricCacheData struct {
	Currency string `json:"currency"`
	Current  string `json:"current"`
	Average  string `json:"average"`
}

func getCurrencyPriceFromDb()[]CurrencyPricCacheData {
	p := QueryAllCaurrencyPrice()
	if len(p) > 0 {
		caches := make([]CurrencyPricCacheData,len(p))
		for i, v := range p{
			caches[i] = CurrencyPricCacheData{
				Currency: v["currency"],
				Current:  v["cur"],
				Average:  v["avg"],
			}
		}
		return caches
	}
	return nil
}
func initCurrencyPriceCache(){
	c,e := ttl(CURRENCY_PRICE_RDS_KEY)
	if e != nil && e != redis.ErrNil {
		logger.Error("ttl rds error",e.Error())
		return
	}

	if c < 0 {
		cps := getCurrencyPriceFromDb()
		if len(cps) > 0 {
			m := make(map[string]string,0)
			for _,v := range cps {
				m[v.Currency] = utils.ToJSON(v)
			}
			hmset(CURRENCY_PRICE_RDS_KEY,m)
			rdsExpire(CURRENCY_PRICE_RDS_KEY,CURRENCY_PRICE_RDS_EXPIRE)
		}
	}
}

func GetCurrencyPrice(currencyPiar string)(bool,*CurrencyPricCacheData){
	initCurrencyPriceCache()
	jsonStr,e := hget(CURRENCY_PRICE_RDS_KEY,currencyPiar)
	if e != nil {
		if e != redis.ErrNil {
			logger.Error("hget rds error",e.Error())
		}
		return false,nil
	}
	res := new(CurrencyPricCacheData)
	utils.FromJson(jsonStr,res)
	return true,res
}