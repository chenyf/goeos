package main

import (
	"fmt"
	"github.com/chenyf/goeos/libraries/appbase"
	"github.com/chenyf/goeos/libraries/fc"
	"github.com/chenyf/goeos/plugins"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	app := appbase.Instance()
	root := fc.AppPath()

	app.RegisterPlugin("chain_plugin", plugins.NewChainPlugin())
	app.RegisterPlugin("http_plugin", plugins.NewHttpPlugin())
	//app.Init("net_plugin")
	//app.Init("producer_plugin")

	fmt.Printf("nodeos version \n")
	fmt.Printf("eosio root is %s\n", root)
	app.Startup()

	wg := &sync.WaitGroup{}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		app.Shutdown()
		wg.Done()
		fmt.Printf("nodeos quit\n")
	}()
	wg.Add(1)
	wg.Wait()

	//app.Exec()
}
