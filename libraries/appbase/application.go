package appbase

import ()

type Application struct {
	Version         uint64
	plugins         map[string]Plugin
	running_plugins []Plugin
}

var application *Application = nil

func Instance() *Application {
	if application == nil {
		application = &Application{
			Version:         0,
			plugins:         map[string]Plugin{},
			running_plugins: []Plugin{},
		}
	}
	return application
}

func (this *Application) RegisterPlugin(name string, plugin Plugin) {
	this.plugins[name] = plugin
}

func (this *Application) Startup() {
	for _, plugin := range this.plugins {
		plugin.Startup()
	}
}

func (this *Application) Exec() {
}

func (this *Application) Shutdown() {
	for _, plugin := range this.running_plugins {
		plugin.Shutdown()
	}
}

func (this *Application) DataDir() string {
	return ""
}
func (this *Application) GetPlugin(name string) Plugin {
	if plugin, ok := this.plugins[name]; ok {
		return plugin
	}
	return nil
}

func (this *Application) ConfigDir() string {
	return ""
}
