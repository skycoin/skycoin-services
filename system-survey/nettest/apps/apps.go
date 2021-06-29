package apps

import (
	"os/exec"
)

func Run(apps []string) map[string]bool {
	if len(apps) == 0 {
		apps = []string{"wget", "git", "go"}
	}

	appInfo := make(map[string]bool)
	for _, app := range apps {
		_, err := exec.LookPath(app)
		appInfo[app] = err == nil
	}

	return appInfo
}
