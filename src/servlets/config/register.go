package config

import (
	"server"
	"servlets/constants"
)

func RegisterHandlers() {
	server.RegisterHandler(constants.CURRENCY_PRICE_QUERY, &currencyPriceHandler{})
	server.RegisterHandler(constants.BATCH_CURRENCY_PRICE_QUERY, &batchCurrencyPriceHandler{})
	server.RegisterHandler(constants.CONFIG_TRANSFER_FEE, &transferFeeHandler{})
	server.RegisterHandler(constants.CONFIG_WITHDRAWAL_FEE, &withdrawalFeeHandler{})
}
