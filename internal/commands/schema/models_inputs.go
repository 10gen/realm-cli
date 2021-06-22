package schema

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"

	"github.com/AlecAivazis/survey/v2"
)

type datamodelsInputs struct {
	cli.ProjectInputs
	Flat      bool
	Language  language
	NoImports bool
	Names     []string
	nameSet   map[string]struct{}
}

func (i *datamodelsInputs) Resolve(profile *user.Profile, ui terminal.UI) error {
	if err := i.ProjectInputs.Resolve(ui, profile.WorkingDirectory, true); err != nil {
		return err
	}

	if i.Language == "" {
		options, typesByOption := make([]string, len(allLanguages)), map[interface{}]language{}
		for i, l := range allLanguages {
			o := languageDisplay(l)
			options[i] = o
			typesByOption[o] = l
		}

		var lang string
		if err := ui.AskOne(&lang, &survey.Select{
			Message: "Select the language you would like to generate data models in",
			Options: options,
		}); err != nil {
			return err
		}
		i.Language = typesByOption[lang]

		if !i.NoImports {
			var noImports bool
			if err := ui.AskOne(&noImports, &survey.Confirm{Message: "Would you like to omit imports?"}); err != nil {
				return err
			}
			i.NoImports = noImports
		}

		if !i.Flat {
			var flat bool
			if err := ui.AskOne(&flat, &survey.Confirm{Message: "Would you like group all generated data models together?"}); err != nil {
				return err
			}
			i.Flat = flat
		}
	}

	i.nameSet = map[string]struct{}{}
	for _, name := range i.Names {
		i.nameSet[name] = struct{}{}
	}

	return nil
}

type language string

const (
	languageEmpty      language = ""
	languageCSharp     language = "c_sharp"
	languageJava       language = "java"
	languageJavascript language = "javascript"
	languageKotlin     language = "kotlin"
	languageObjectiveC language = "objective_c"
	languageSwift      language = "swift"
	languageTypescript language = "typescript"
)

var (
	allLanguages = []language{
		languageCSharp,
		languageJava,
		languageJavascript,
		languageKotlin,
		languageObjectiveC,
		languageSwift,
		languageTypescript,
	}

	allLanguageAliases = map[string]language{
		"c-sharp":     languageCSharp,
		"csharp":      languageCSharp,
		"c#":          languageCSharp,
		"js":          languageJavascript,
		"objective-c": languageObjectiveC,
		"objectivec":  languageObjectiveC,
		"obj_c":       languageObjectiveC,
		"obj-c":       languageObjectiveC,
		"objc":        languageObjectiveC,
		"ts":          languageTypescript,
	}
)

func (l language) String() string { return string(l) }

func (l language) Type() string { return flags.TypeString }

func (l *language) Set(val string) error {
	lang, ok := validateLanguage(val)
	if !ok {
		languages := make([]string, 0, len(allLanguages))
		for _, l := range allLanguages {
			languages = append(languages, string(l))
		}
		return fmt.Errorf("'%s' is not a recognized language, instead try: %s", val, strings.Join(languages, ", "))
	}

	*l = lang
	return nil
}

func validateLanguage(val string) (language, bool) {
	if val == "" {
		return languageEmpty, true
	}

	v := strings.ToLower(val)

	for _, l := range allLanguages {
		if l == language(v) {
			return l, true
		}
	}

	if l, ok := allLanguageAliases[v]; ok {
		return l, true
	}

	return language(""), false
}
