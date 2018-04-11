package plugins

import (
	"fmt"
	"net/http"
)

type HttpPlugin struct {
	name string
	addr string
}

func NewHttpPlugin() *HttpPlugin {
	return &HttpPlugin{
		name: "http_plugin",
	}
}

func (this *HttpPlugin) Init() {
}

func (this *HttpPlugin) Startup() error {
	fmt.Printf("%s startup\n", this.name)
	err := http.ListenAndServe(this.addr, nil)
	return err
}

func (this *HttpPlugin) Shutdown() {
	fmt.Printf("%s shutdown\n", this.name)
}

func (this *HttpPlugin) Name() string {
	return this.name
}
