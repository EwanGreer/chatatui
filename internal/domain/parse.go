package domain

import (
	"fmt"
	"strings"
)

func parseAtParts(s, typeName, leftField string) (left, right string, err error) {
	left, right, found := strings.Cut(s, "@")
	switch {
	case !found:
		return "", "", fmt.Errorf("invalid %s %q: missing @", typeName, s)
	case left == "":
		return "", "", fmt.Errorf("invalid %s %q: empty %s", typeName, s, leftField)
	case right == "":
		return "", "", fmt.Errorf("invalid %s %q: empty domain", typeName, s)
	case strings.Contains(right, "@"):
		return "", "", fmt.Errorf("invalid %s %q: multiple @ signs", typeName, s)
	}
	return left, right, nil
}
