package plugins

import ()

type NetPlugin struct {
	addr string
}

func (this *NetPlugin) Init() {
}

func (this *NetPlugin) Startup() error {
	if this.acceptor {
	}

	this.start_monitors()
	for _, seed := range this.supplied_peers {
		connect(seed)
	}
	return nil
}

func (this *NetPlugin) Shutdown() {
}

func (this *NetPlugin) Name() string {
	return "network plugin"
}
