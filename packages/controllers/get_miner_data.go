package controllers

import (
	"encoding/json"
	"github.com/c-darwin/dcoin-go/packages/utils"

	"math"
)

func (c *Controller) GetMinerData() (string, error) {

	c.r.ParseForm()

	secs := float64(3600 * 24 * 365)

	userId := utils.StrToInt64(c.r.FormValue("userId"))
	if !utils.CheckInputData(userId, "int") {
		return `{"result":"incorrect userId"}`, nil
	}

	minersData, err := c.OneRow("SELECT * FROM miners_data WHERE user_id  =  ?", userId).String()
	if err != nil {
		return "", err
	}

	// получим ID майнеров, у которых лежат фото нужного нам юзера
	minersIds := utils.GetMinersKeepers(minersData["photo_block_id"], minersData["photo_max_miner_id"], minersData["miners_keepers"], false)
	hosts, err := c.GetList("SELECT http_host as host FROM miners_data WHERE miner_id IN (" + utils.JoinInts(minersIds, ",") + ")").String()
	if err != nil {
		return "", err
	}

	currencyList, err := c.GetCurrencyList(false)
	if err != nil {
		return "", err
	}

	_, _, promisedAmountListGen, err := c.GetPromisedAmounts(userId, c.Variables.Int64["cash_request_time"])
	var data utils.DCAmounts
	if promisedAmountListGen[72].Amount > 0 {
		data = promisedAmountListGen[72]
	} else if promisedAmountListGen[23].Amount > 0 {
		data = promisedAmountListGen[23]
	} else {
		data = utils.DCAmounts{}
	}

	promisedAmounts := ""
	prognosis := make(map[int64]float64)
	if data.Amount > 1 {
		promisedAmounts += utils.Float64ToStr(utils.Round(data.Amount, 0)) + " " + currencyList[(data.CurrencyId)] + "<br>"
		prognosis[int64(data.CurrencyId)] += (math.Pow(1+data.PctSec, secs) - 1) * data.Amount
	}

	if len(promisedAmounts) > 0 {
		promisedAmounts = "<strong>" + promisedAmounts[:len(promisedAmounts)-4] + "</strong><br>" + c.Lang["promised"] + "<hr>"
	}

	/*
	 * На кошельках
	 * */

	balances, err := c.GetBalances(userId)
	if err != nil {
		return "", err
	}
	walletsByCurrency := make(map[int]utils.DCAmounts)
	for _, data := range balances {
		walletsByCurrency[int(data.CurrencyId)] = data
	}
	if walletsByCurrency[72].Amount > 0 {
		data = walletsByCurrency[72]
	} else if walletsByCurrency[23].Amount > 0 {
		data = walletsByCurrency[23]
	} else {
		data = utils.DCAmounts{}
	}

	wallets := ""
	var countersIds []string
	var pctSec float64
	if data.Amount > 0 {
		counterId := "map-" + utils.Int64ToStr(userId) + "-" + utils.Int64ToStr(data.CurrencyId)
		countersIds = append(countersIds, counterId)
		wallets = "<span class='dc_amount' id='" + counterId + "'>" + utils.Float64ToStr(data.Amount) + "</span> d" + currencyList[(data.CurrencyId)] + "<br>"
		// прогноз
		prognosis[int64(data.CurrencyId)] += (math.Pow(1+data.PctSec, secs) - 1) * data.Amount
		pctSec = data.PctSec
	}

	if len(wallets) > 0 {
		wallets = wallets[:len(wallets)-4] + "<br>" + c.Lang["on_the_account"] + "<hr>"
	}

	/*
	 * Годовой прогноз
	 * */
	prognosisHtml := ""
	for currencyId, amount := range prognosis {
		if amount < 0.01 {
			continue
		} else if amount < 1 {
			amount = utils.Round(amount, 2)
		} else {
			amount = amount
		}
		prognosisHtml += "<span class='amount_1year'>" + utils.Float64ToStr(amount) + " d" + currencyList[(currencyId)] + "</span><br>"
	}
	if len(prognosisHtml) > 0 {
		prognosisHtml = prognosisHtml[:len(prognosisHtml)-4] + "<br> " + c.Lang["profit_forecast"] + " " + c.Lang["after_1_year"]
	}

	result_ := minersDataType{Hosts: hosts, Lnglat: map[string]string{"lng": minersData["longitude"], "lat": minersData["latitude"]}, Html: promisedAmounts + wallets + "<div style=\"clear:both\"></div>" + prognosisHtml + "</p>", Counters: countersIds, PctSec: pctSec}
	log.Debug("result_", result_)
	result, err := json.Marshal(result_)
	if err != nil {
		return "", err
	}
	log.Debug(string(result))
	return string(result), nil
}

type minersDataType struct {
	Hosts    []string          `json:"hosts"`
	Lnglat   map[string]string `json:"lnglat"`
	Html     string            `json:"html"`
	Counters []string          `json:"counters"`
	PctSec   float64           `json:"pct_sec"`
}
