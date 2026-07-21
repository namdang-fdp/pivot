package cli

import "strings"

// shellQuote returns value as one safety copyable shell argument
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
