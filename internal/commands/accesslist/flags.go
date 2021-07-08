package accesslist

import "github.com/10gen/realm-cli/internal/utils/flags"

func ipFlag(value *string, description string) flags.StringFlag {
	return flags.StringFlag{
		Value: value,
		Meta: flags.Meta{
			Name: "ip",
			Usage: flags.Usage{
				Description: description,
			},
		},
	}
}

func commentFlag(value *string, description string) flags.StringFlag {
	return flags.StringFlag{
		Value: value,
		Meta: flags.Meta{
			Name: "comment",
			Usage: flags.Usage{
				Description: description,
			},
		},
	}
}
