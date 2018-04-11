package appbase

type Plugin interface {
	Startup() error
	Shutdown()
}
