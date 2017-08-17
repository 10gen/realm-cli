package commands

import (
	"fmt"
	"strings"
)

func subCommandUsageFormat(helps []struct{ Name, Help string }) string {
	var max string
	for _, st := range helps {
		if len(st.Name) > len(max) {
			max = st.Name
		}
	}

	var lines []string
	for _, st := range helps {
		helpLines := strings.Split(st.Help, "\n")
		if len(st.Help) == 0 {
			lines = append(lines, fmt.Sprintf("  %s", st.Name))
		} else {
			filler := strings.Repeat(" ", len(max)-len(st.Name))
			lines = append(lines, fmt.Sprintf("  %s%s\t%s", st.Name, filler, helpLines[0]))
			prefix := strings.Repeat(" ", len(max)+2)
			for _, line := range helpLines[1:] {
				lines = append(lines, fmt.Sprintf("%s\t%s", prefix, line))
			}
		}
	}
	return fmt.Sprintf("%s\n", strings.Join(lines, "\n"))
}
