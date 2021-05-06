package schema

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/10gen/realm-cli/internal/utils/test/mock"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/Netflix/go-expect"
	"github.com/google/go-cmp/cmp"
)

func TestSchemaModelsInputsResolve(t *testing.T) {
	t.Run("should prompt user for no information if language is set", func(t *testing.T) {
		assert.RegisterOpts(reflect.TypeOf(datamodelsInputs{}), cmp.AllowUnexported(datamodelsInputs{}))
		profile := mock.NewProfile(t)

		_, ui := mock.NewUI()

		i := datamodelsInputs{Language: languageTypescript}

		assert.Nil(t, i.Resolve(profile, ui))
		assert.Equal(t, datamodelsInputs{Language: languageTypescript, nameSet: map[string]struct{}{}}, i)
	})

	for _, tc := range []struct {
		description string
		inputs      datamodelsInputs
		procedure   func(c *expect.Console)
		expected    datamodelsInputs
	}{
		{
			description: "should prompt user for all information if none provided",
			procedure: func(c *expect.Console) {
				c.ExpectString("Select the language you would like to generate data models in")
				c.Send(string(terminal.KeyArrowDown))
				c.SendLine("") // select java
				c.ExpectString("Would you like to omit imports?")
				c.SendLine("y")
				c.ExpectString("Would you like group all generated data models together?")
				c.SendLine("")
				c.ExpectEOF()
			},
			expected: datamodelsInputs{
				Language:  languageJava,
				NoImports: true,
				Flat:      false,
			},
		},
		{
			description: "should not prompt for no imports if already set",
			inputs:      datamodelsInputs{NoImports: true},
			procedure: func(c *expect.Console) {
				c.ExpectString("Select the language you would like to generate data models in")
				c.Send(string(terminal.KeyArrowDown))
				c.Send(string(terminal.KeyArrowDown))
				c.SendLine("") // select javascript
				c.ExpectString("Would you like group all generated data models together?")
				c.SendLine("y")
				c.ExpectEOF()
			},
			expected: datamodelsInputs{
				Language:  languageJavascript,
				NoImports: true,
				Flat:      true,
			},
		},
		{
			description: "should not prompt for flat if already set",
			inputs:      datamodelsInputs{Flat: true},
			procedure: func(c *expect.Console) {
				c.ExpectString("Select the language you would like to generate data models in")
				c.SendLine("ty") // select typescript
				c.ExpectString("Would you like to omit imports?")
				c.SendLine("y")
				c.ExpectEOF()
			},
			expected: datamodelsInputs{
				Language:  languageTypescript,
				NoImports: true,
				Flat:      true,
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			profile := mock.NewProfile(t)

			_, console, _, ui, err := mock.NewVT10XConsole()
			assert.Nil(t, err)
			defer console.Close()

			doneCh := make(chan struct{})
			go func() {
				defer close(doneCh)
				tc.procedure(console)
			}()

			tc.inputs.App = "some-app" // avoid app resolution
			assert.Nil(t, tc.inputs.Resolve(profile, ui))

			console.Tty().Close() // flush the writers
			<-doneCh

			assert.Equal(t, tc.expected.Flat, tc.inputs.Flat)
			assert.Equal(t, tc.expected.Language, tc.inputs.Language)
			assert.Equal(t, tc.expected.NoImports, tc.inputs.NoImports)
			assert.Equal(t, tc.expected.Names, tc.inputs.Names)
		})
	}
}

func TestLanguageValidate(t *testing.T) {
	t.Run("should not return ok with an invalid language", func(t *testing.T) {
		l, ok := validateLanguage("eggcorn")
		assert.False(t, ok, "expected invalid language")
		assert.Equal(t, language(""), l)
	})

	for _, tc := range []struct {
		val      string
		expected language
	}{
		{"", languageEmpty},
		{string(languageCSharp), languageCSharp},
		{string(languageJava), languageJava},
		{string(languageJavascript), languageJavascript},
		{string(languageKotlin), languageKotlin},
		{string(languageObjectiveC), languageObjectiveC},
		{string(languageSwift), languageSwift},
		{string(languageTypescript), languageTypescript},
		{"c-sharp", languageCSharp},
		{"csharp", languageCSharp},
		{"c#", languageCSharp},
		{"js", languageJavascript},
		{"objective-c", languageObjectiveC},
		{"objectivec", languageObjectiveC},
		{"obj_c", languageObjectiveC},
		{"obj-c", languageObjectiveC},
		{"objc", languageObjectiveC},
		{"ts", languageTypescript},
	} {
		t.Run(fmt.Sprintf("should identify %s as a valid language", tc.val), func(t *testing.T) {
			l, ok := validateLanguage(tc.val)
			assert.True(t, ok, "expected valid language")
			assert.Equal(t, tc.expected, l)
		})
	}

	t.Run("should identify languages while ignoring casing", func(t *testing.T) {
		l, ok := validateLanguage("Objective-C")
		assert.True(t, ok, "expected valid language")
		assert.Equal(t, languageObjectiveC, l)
	})
}

func TestLanguageSet(t *testing.T) {
	type holder struct {
		l *language
	}

	newHolder := func() holder {
		var l language
		return holder{&l}
	}

	t.Run("should set the language value", func(t *testing.T) {
		tc := newHolder()
		assert.Equal(t, language(""), *tc.l)

		assert.Nil(t, tc.l.Set("c_sharp"))
		assert.Equal(t, languageCSharp, *tc.l)

		assert.Nil(t, tc.l.Set("objc"))
		assert.Equal(t, languageObjectiveC, *tc.l)
	})

	t.Run("should return an error when setting an invalid error", func(t *testing.T) {
		tc := newHolder()

		expectedErr := errors.New("'eggcorn' is not a recognized language, instead try: c_sharp, java, javascript, kotlin, objective_c, swift, typescript")
		assert.Equal(t, expectedErr, tc.l.Set("eggcorn"))
	})
}
