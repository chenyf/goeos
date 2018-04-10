package appbase

type Plugin struct {
	Startup func() bool
}
