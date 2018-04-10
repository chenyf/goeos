package plugins

import (
	"net/http"
)

type HttpPlugin struct {
	addr string
}

func (this *HttpPlugin) Init() {
}

func (this *HttpPlugin) Startup() error {
	err := http.ListenAndServe(this.addr, nil)
	return err
}

func (this *HttpPlugin) Shutdown() {
}

func (this *HttpPlugin) Name() string {
	return "http plugin"
}
