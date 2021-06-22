package secrets

import "github.com/10gen/realm-cli/internal/utils/flags"

const (
	flagSecret      = "secret"
	flagSecretShort = "s"
)

func nameFlag(value *string, description string) flags.StringFlag {
	return flags.StringFlag{
		Value: value,
		Meta: flags.Meta{
			Name:      "name",
			Shorthand: "n",
			Usage: flags.Usage{
				Description: description,
			},
		},
	}
}

func valueFlag(value *string, description string) flags.StringFlag {
	return flags.StringFlag{
		Value: value,
		Meta: flags.Meta{
			Name:      "value",
			Shorthand: "v",
			Usage: flags.Usage{
				Description: description,
			},
		},
	}
}
