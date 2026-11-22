package health

import "embed"

//go:embed monitor.go simulator.go sequential.go concurrent.go
var sourceFiles embed.FS

func Source(component string) (string, bool) {
	switch component {
	case "monitor", "simulator", "sequential", "concurrent":
		data, err := sourceFiles.ReadFile(component + ".go")

		if err != nil {
			return "", false
		}

		return string(data), true

	default:
		return "", false
	}
}
