package appbase

type Plugin interface {
	Startup() bool
}
