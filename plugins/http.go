package plugins

import (
	"fmt"
	"net/http"
)

type HttpPlugin struct {
	addr string
}

func NewHttpPlugin() *HttpPlugin {
	return &HttpPlugin{}
}

func (this *HttpPlugin) Init() {
}

func (this *HttpPlugin) Startup() error {
	fmt.Printf("http plugin startup\n")
	err := http.ListenAndServe(this.addr, nil)
	return err
}

func (this *HttpPlugin) Shutdown() {
}

func (this *HttpPlugin) Name() string {
	return "http plugin"
}
