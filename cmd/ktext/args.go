package main

import "strings"

// splitArgs separates flag arguments from positional arguments, allowing flags
// to appear before or after positional args. boolFlags names flags that take no
// value (so the next token is not consumed as their value).
func splitArgs(args []string, boolFlags map[string]bool) (flags, positional []string) {
	i := 0
	for i < len(args) {
		a := args[i]
		if !strings.HasPrefix(a, "-") {
			positional = append(positional, a)
			i++
			continue
		}
		flags = append(flags, a)
		name := strings.TrimLeft(a, "-")
		// -flag=value embeds the value; no lookahead needed.
		if strings.Contains(name, "=") {
			i++
			continue
		}
		// Boolean flags consume no following token.
		if boolFlags[name] {
			i++
			continue
		}
		// String flags: consume the next token as the value if it doesn't look like a flag.
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			i++
			flags = append(flags, args[i])
		}
		i++
	}
	return
}
