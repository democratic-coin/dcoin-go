// notifications
package dcparser

import (
	"fmt"
	"encoding/json"
	"github.com/democratic-coin/dcoin-go/packages/utils"
)

func  (p *Parser) isNotify() bool {
	if val,ok := p.ConfigIni["notify"]; ok && val == `1`{
		return true
	}
	return false
}

func  (p *Parser) nfyRollback( blockId int64 ) {
	if !p.isNotify() {
		return
	}
	p.ExecSql( `delete from notifications where block_id=?`, blockId )
}


func (p *Parser) insertNotify( userId int64, cmdId int, params string) {
	p.ExecSql("insert into notifications (user_id, block_id, cmd_id, params) VALUES (?, ?, ?, ?)", 
	          userId, p.BlockData.BlockId, cmdId, params )
}

func  (p *Parser) nfyRefReady( userId int64, refId int64 ) {
	if !p.isNotify() {
		return
	}
	p.insertNotify( userId, utils.ECMD_REFREADY, fmt.Sprintf( `{"refid": "%d"}`, refId ))
}

func  (p *Parser) nfyStatus( userId int64, status string ) {
	if !p.isNotify() {
		return
	}
	p.insertNotify( userId, utils.ECMD_CHANGESTAT, fmt.Sprintf( `{"status": "%s"}`, status ))
}

func  (p *Parser) nfySent( userId int64, tns *utils.TypeNfySent ) {
	if !p.isNotify() {
		return
	}
	params,err := json.Marshal( tns ) 
	if err != nil {
		params = []byte(fmt.Sprintf( `{"error": "%s"}`, err ))
	}
	p.insertNotify( userId, utils.ECMD_DCSENT, string(params))
}

func  (p *Parser) nfyCame( userId int64, tnc *utils.TypeNfyCame ) {
	if !p.isNotify() {
		return
	}
	params,err := json.Marshal( tnc ) 
	if err != nil {
		params = []byte(fmt.Sprintf( `{"error": "%s"}`, err ))
	}
	p.insertNotify( userId, utils.ECMD_DCCAME, string(params))
}
