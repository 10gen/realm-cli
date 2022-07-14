module github.com/10gen/realm-cli

go 1.14

require (
	github.com/AlecAivazis/survey/v2 v2.2.3
	github.com/Netflix/go-expect v0.0.0-20180615182759-c93bf25de8e8
	github.com/blang/semver v3.5.1+incompatible
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/briandowns/spinner v1.12.0
	github.com/edaniels/digest v0.0.0-20170923160545-b81e9c4ee11c
	github.com/edaniels/golinters v0.0.3
	github.com/fatih/color v1.10.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golangci/golangci-lint v1.32.2
	github.com/google/go-cmp v0.5.2
	github.com/hinshun/vt10x v0.0.0-20180616224451-1954e6464174
	github.com/iancoleman/orderedmap v0.1.0
	github.com/kr/pretty v0.2.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/segmentio/backo-go v0.0.0-20200129164019-23eae7c10bd3 // indirect
	github.com/spf13/afero v1.1.2
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v0.0.0-20180714160509-73f8eece6fdc // indirect
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	go.mongodb.org/mongo-driver v1.5.1
	gopkg.in/segmentio/analytics-go.v3 v3.1.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/edaniels/golinters => github.com/mongodb-forks/golinters v0.0.4
