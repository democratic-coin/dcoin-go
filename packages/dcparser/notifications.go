// notifications
package dcparser

import (
	"fmt"
	"github.com/democratic-coin/dcoin-go/packages/utils"
)

func  (p *Parser) isNotify() bool {
	if val,ok := p.ConfigIni["notify"]; ok && val == `1`{
		return true
	}
	return false
}

func (p *Parser) insertNotify( userId int64, cmdId int, params string) {
	p.ExecSql("insert into notifications (user_id, block_id, cmd_id, params) VALUES (?, ?, ?, ?)", 
	          userId, p.BlockData.BlockId, cmdId, params )
}

func  (p *Parser) nfyStatus( userId int64, status string ) {
	if !p.isNotify() {
		return
	}
	p.insertNotify( userId, utils.ECMD_CHANGESTAT, fmt.Sprintf( `{"status": "%s"}`, status ))
}
