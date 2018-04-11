package plugins

import (
	"fmt"
	"net/http"
)

type ProducerPlugin struct {
	name string
	addr string
}

func NewProducerPlugin() *ProducerPlugin {
	return &ProducerPlugin{
		name: "producer_plugin",
	}
}

func (this *ProducerPlugin) Init() {
}

func (this *ProducerPlugin) Startup() error {
	fmt.Printf("%s startup\n", this.name)
	err := http.ListenAndServe(this.addr, nil)
	return err
}

func (this *ProducerPlugin) Shutdown() {
	fmt.Printf("%s shutdown\n", this.name)
}

func (this *ProducerPlugin) Name() string {
	return this.name
}
