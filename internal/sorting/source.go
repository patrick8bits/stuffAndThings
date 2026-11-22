package sorting

import "embed"

//go:embed bubble.go insertion.go quick.go
var sourceFiles embed.FS

func Source(algorithm string) (string, bool) {
	switch algorithm {
	case "bubble", "insertion", "quick":
		data, err := sourceFiles.ReadFile(algorithm + ".go")

		if err != nil {
			return "", false
		}

		return string(data), true

	default:
		return "", false
	}
}
