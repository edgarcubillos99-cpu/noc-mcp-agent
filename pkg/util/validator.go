package util

import "regexp"

var validTargetRegex = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

// IsValidTarget previene inyecciones de comandos validando el input del LLM
func IsValidTarget(t string) bool {
	return validTargetRegex.MatchString(t)
}
