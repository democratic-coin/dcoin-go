package controllers

import (
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"net/http"
	"fmt"
)


func IndexGetPool(w http.ResponseWriter, r *http.Request) {

	fmt.Println("IndexGetPool");
	if utils.DB != nil && utils.DB.DB != nil {

		var err error
		var poolHttpHost []byte
		getUserId := utils.StrToInt64(r.FormValue("user_id"))
		if getUserId == 0 {
			variables, err := utils.DB.GetAllVariables()
			poolHttpHost, err = utils.DB.Single(`SELECT http_host FROM miners_data WHERE i_am_pool = 1 AND pool_count_users < ?`, variables.Int64["max_pool_users"]).Bytes()
			if err != nil {
				log.Error("%v", err)
			}
		} else {
			poolHttpHost, err = utils.DB.Single("SELECT CASE WHEN m.pool_user_id > 0 then (SELECT http_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE http_host end as http_host FROM miners_data as m WHERE m.user_id = ?", getUserId).Bytes()
			if err != nil {
				log.Error("%v", err)
			}
		}
		if _, err := w.Write(poolHttpHost); err != nil {
			log.Error("%v", err)
		}

	}
}