package versionbundle

type App struct {
	App              string `yaml:"app"`
	ComponentVersion string `yaml:"componentVersion"`
	Version          string `yaml:"version"`
}

func CopyApps(apps []App) []App {
	appList := make([]App, len(apps))
	copy(appList, apps)
	return appList
}
