package ui

import (
	"fmt"
	"os"
)

// Yes disables "are you sure?"-style prompts.
var Yes bool

// Ask asks the user a question unless Yes is true.
func Ask(query string) bool {
	if Yes {
		return true
	}
	fmt.Fprintf(os.Stderr, "%s [y/n] ", query)
	for {
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if response[0] == 'n' {
			return false
		}
		if response[0] == 'y' {
			return true
		}
		fmt.Fprint(os.Stderr, "Could not understand response, try again. [y/n] ")
	}
}
