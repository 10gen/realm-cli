package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/10gen/stitch-cli/ui"
)

type permissions string

const (
	permissionsAny permissions = ""
	permissionsRW              = "rw"
	permissionsR               = "r"
)

type permissionedItem struct {
	perms permissions
	name  string
}

func printPermissionedItems(items []permissionedItem, hidePerms bool) {
	if hidePerms {
		for _, item := range items {
			fmt.Println(item.name)
		}
		return
	}
	for _, item := range items {
		padding := strings.Repeat(" ", 2-len(item.perms))
		p := ui.Color(ui.Permissions, string(item.perms))
		fmt.Printf("%s%s\t%s\n", p, padding, item.name)
	}
}

func printGroupedPermissionedItems(items map[string][]permissionedItem, hidePerms bool) {
	sortedGroupNames := make([]string, 0, len(items))
	for groupName := range items {
		sortedGroupNames = append(sortedGroupNames, groupName)
	}
	sort.Strings(sortedGroupNames)
	for _, groupName := range sortedGroupNames {
		if len(items[groupName]) == 0 {
			continue
		}
		fmt.Printf("%s:\n", groupName)
		if hidePerms {
			for _, item := range items[groupName] {
				fmt.Printf("\t%s\n", item.name)
			}
		} else {
			for _, item := range items[groupName] {
				padding := strings.Repeat(" ", 2-len(item.perms))
				p := ui.Color(ui.Permissions, string(item.perms))
				fmt.Printf("\t%s%s\t%s\n", p, padding, item.name)
			}
		}
	}
}

type kv struct {
	key, value string
	values     []string
	valuePairs [][2]string
}

func printKV(items []kv) {
	var max int
	for _, item := range items {
		if len(item.key) > max {
			max = len(item.key)
		}
	}

	for _, item := range items {
		if item.value != "" {
			padding := strings.Repeat(" ", max-len(item.key))
			fmt.Printf("%s:%s\t%s\n", item.key, padding, item.value)
			continue
		}
		if item.values != nil {
			fmt.Printf("%s:\n", item.key)
			for _, v := range item.values {
				fmt.Printf("\t%s\n", v)
			}
			continue
		}
		if item.valuePairs != nil {
			fmt.Printf("%s:\n", item.key)
			var max2 int
			for _, ss := range item.valuePairs {
				if len(ss[0]) > max2 {
					max2 = len(ss[0])
				}
			}
			for _, ss := range item.valuePairs {
				padding := strings.Repeat(" ", max2-len(ss[0]))
				fmt.Printf("\t%s%s\t%s\n", ss[0], padding, ss[1])
			}
			continue
		}
		fmt.Printf("%s:\n", item.key)
	}
}

func printSingleKV(item kv) {
	if item.value != "" {
		fmt.Println(item.value)
		return
	}
	if item.values != nil {
		for _, v := range item.values {
			fmt.Println(v)
		}
		return
	}
	if item.valuePairs != nil {
		var max int
		for _, ss := range item.valuePairs {
			if len(ss[0]) > max {
				max = len(ss[0])
			}
		}
		for _, ss := range item.valuePairs {
			padding := strings.Repeat(" ", max-len(ss[0]))
			fmt.Printf("%s%s\t%s\n", ss[0], padding, ss[1])
		}
		return
	}
}

func validGroup(group string) bool {
	return len(group) > 2 // TODO
}
