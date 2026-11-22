package cache

import "embed"

//go:embed lru.go origin.go service.go workload.go
var sourceFiles embed.FS

// Source returns the embedded Go source for an allowed component.
func Source(component string) (string, bool) {
	switch component {
	case "lru", "origin", "service", "workload":
		data, err := sourceFiles.ReadFile(component + ".go")

		if err != nil {
			return "", false
		}

		return string(data), true

	default:
		return "", false
	}
}
