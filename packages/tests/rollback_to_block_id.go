package main

import (
	"fmt"
//	"github.com/democratic-coin/dcoin-go/packages/utils"
	"github.com/democratic-coin/dcoin-go/packages/tests_utils"
	"github.com/democratic-coin/dcoin-go/packages/dcparser"
)

func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	db := tests_utils.DbConn()
	parser := new(dcparser.Parser)
	parser.DCDB = db
	err := parser.RollbackToBlockId(261950)
	if err!=nil {
		fmt.Println(err)
	}

}
