package plugins

import (
	"net/http"
)

type ProducerPlugin struct {
	addr string
}

func (this *ProducerPlugin) Init() {
}

func (this *ProducerPlugin) Startup() error {
	err := http.ListenAndServe(this.addr, nil)
	return err
}

func (this *ProducerPlugin) Shutdown() {
}

func (this *ProducerPlugin) Name() string {
	return "producer plugin"
}
