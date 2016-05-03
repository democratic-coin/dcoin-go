package main

import (
	"fmt"
	"github.com/democratic-coin/dcoin-go/packages/utils"
	"github.com/democratic-coin/dcoin-go/packages/tests_utils"
	"github.com/democratic-coin/dcoin-go/packages/dcparser"
)

func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	d := tests_utils.DbConn()
	p := new(dcparser.Parser)
	p.DCDB = d
	var myTestBlockBody []byte
	transactionsTestblock, err := d.GetAll("SELECT data FROM transactions_testblock ORDER BY id ASC", -1)
	if err != nil {
		fmt.Println(utils.ErrInfo(err))
	}
	for _, data := range transactionsTestblock {
		fmt.Printf("%x\n", data["data"])
		myTestBlockBody = append(myTestBlockBody, utils.EncodeLengthPlusData([]byte(data["data"]))...)
	}
	fmt.Println(utils.BinToHex(myTestBlockBody))

	if len(myTestBlockBody) > 0 {
		fmt.Printf("%x\n", myTestBlockBody)
		p.BinaryData = append(utils.DecToBin(0, 1), myTestBlockBody...)
		err = p.ParseDataGate(true)
		if err != nil {
			fmt.Println(utils.ErrInfo(err))
		}
	}
}
