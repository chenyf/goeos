package plugins

import (
	"fmt"
	"net/http"
)

type ProducerPlugin struct {
	addr string
}

func (this *ProducerPlugin) Init() {
}

func (this *ProducerPlugin) Startup() error {
	fmt.Printf("producer plugin startup\n")
	err := http.ListenAndServe(this.addr, nil)
	return err
}

func (this *ProducerPlugin) Shutdown() {
}

func (this *ProducerPlugin) Name() string {
	return "producer plugin"
}
