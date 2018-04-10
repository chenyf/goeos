package main

import (
	"fmt"
	"github.com/chenyf/goeos/libraries/appbase"
	"github.com/chenyf/goeos/libraries/fc"
)

func main() {
	app := appbase.Instance()
	root := fc.AppPath()

	app.Init("chain_plugin")
	app.Init("http_plugin")
	app.Init("net_plugin")
	app.Init("producer_plugin")

	fmt.Printf("nodeos version ")
	fmt.Printf("eosio root is %s\n", root)
	app.Startup()
	app.Exec()
}
