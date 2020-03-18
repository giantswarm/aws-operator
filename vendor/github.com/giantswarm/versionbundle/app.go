package versionbundle

type App struct {
	App              string `yaml:"app"`
	ComponentVersion string `yaml:"componentVersion"`
	Version          string `yaml:"version"`
}

func (a App) AppID() string {
	return a.App + ":" + a.Version
}

func CopyApps(apps []App) []App {
	appList := make([]App, len(apps))
	copy(appList, apps)
	return appList
}
