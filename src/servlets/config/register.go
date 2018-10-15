package config

import (
	"server"
	"servlets/constants"
)

func RegisterHandlers() {
	server.RegisterHandler(constants.CURRENCY_PRICE_QUERY, &currencyPriceHandler{})
}
