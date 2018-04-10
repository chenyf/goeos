package appbase

import ()

type Application struct {
	Version uint64
	plugins []Plugin
}

func Instance() *Application {
	return nil
}

func (this *Application) Init(plugin string) {
}

func (this *Application) Startup() {
	for _, plugin := range this.plugins {
		plugin.Startup()
	}
}

func (this *Application) Exec() {
}

func (this *Application) Shutdown() {
}

func (this *Application) DataDir() string {
	return ""
}

func (this *Application) ConfigDir() string {
	return ""
}
