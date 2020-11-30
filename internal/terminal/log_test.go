package terminal

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/google/go-cmp/cmp"
)

func TestLogConstructor(t *testing.T) {
	assert.RegisterOpts(reflect.TypeOf(jsonDocument{}), cmp.AllowUnexported(jsonDocument{}))
	assert.RegisterOpts(reflect.TypeOf(titledJSONDocument{}), cmp.AllowUnexported(titledJSONDocument{}), cmp.AllowUnexported(jsonDocument{}))

	for _, tc := range []struct {
		ctor          string
		log           Log
		expectedLevel LogLevel
		expectedData  LogData
	}{
		{
			ctor:          "NewTextLog",
			log:           NewTextLog("oh yeah"),
			expectedLevel: LogLevelInfo,
			expectedData:  textMessage("oh yeah"),
		},
		{
			ctor:          "NewJSONLog",
			log:           NewJSONLog(map[string]interface{}{"a": "ayyy"}),
			expectedLevel: LogLevelInfo,
			expectedData:  jsonDocument{map[string]interface{}{"a": "ayyy"}},
		},
		{
			ctor:          "NewTitledJSONLog",
			log:           NewTitledJSONLog("Test Title", map[string]interface{}{"a": "ayyy"}),
			expectedLevel: LogLevelInfo,
			expectedData:  titledJSONDocument{"Test Title", jsonDocument{map[string]interface{}{"a": "ayyy"}}},
		},
		{
			ctor:          "NewErrorLog",
			log:           NewErrorLog(errors.New("oh noz")),
			expectedLevel: LogLevelError,
			expectedData:  errorMessage{errors.New("oh noz")},
		},
	} {
		t.Run(fmt.Sprintf("%s should create the expected Log", tc.ctor), func(t *testing.T) {
			time.Sleep(1 * time.Millisecond) // force tick
			assert.True(t, time.Now().After(tc.log.Time), "now should be later than the log's timestamp")
			assert.Equal(t, tc.expectedLevel, tc.log.Level)
			assert.Equal(t, tc.expectedData, tc.log.Data)
		})
	}
}

func TestLogMessage(t *testing.T) {
	for _, tc := range []struct {
		level           LogLevel
		data            LogData
		expectedOutputs map[OutputFormat]string
	}{
		{
			level: LogLevelInfo,
			data:  textMessage("this is a test log"),
			expectedOutputs: map[OutputFormat]string{
				OutputFormatText: "07:54:00 UTC INFO  this is a test log",
				OutputFormatJSON: `{"time":"1989-06-22T07:54:00Z","level":"info","message":"this is a test log"}`,
			},
		},
		{
			level: LogLevelInfo,
			data:  jsonDocument{map[string]interface{}{"a": true, "b": 1, "c": "sea"}},
			expectedOutputs: map[OutputFormat]string{
				OutputFormatText: `07:54:00 UTC INFO  {
  "a": true,
  "b": 1,
  "c": "sea"
}`,
				OutputFormatJSON: `{"time":"1989-06-22T07:54:00Z","level":"info","doc":{"a":true,"b":1,"c":"sea"}}`,
			},
		},
		{
			level: LogLevelInfo,
			data:  titledJSONDocument{"Test Title", jsonDocument{map[string]interface{}{"a": true, "b": 1, "c": "sea"}}},
			expectedOutputs: map[OutputFormat]string{
				OutputFormatText: `07:54:00 UTC INFO  Test Title
---
{
  "a": true,
  "b": 1,
  "c": "sea"
}`,
				OutputFormatJSON: `{"time":"1989-06-22T07:54:00Z","level":"info","title":"Test Title","doc":{"a":true,"b":1,"c":"sea"}}`,
			},
		},
		{
			level: LogLevelError,
			data:  errorMessage{errors.New("something bad happened")},
			expectedOutputs: map[OutputFormat]string{
				OutputFormatText: "07:54:00 UTC ERROR something bad happened",
				OutputFormatJSON: `{"time":"1989-06-22T07:54:00Z","level":"error","err":"something bad happened"}`,
			},
		},
	} {
		for outputFormat, expectedOutput := range tc.expectedOutputs {
			t.Run(fmt.Sprintf("With %s output format, %T should print the expected output", outputFormat, tc.data), func(t *testing.T) {
				log := Log{
					tc.level,
					time.Date(1989, 6, 22, 7, 54, 0, 0, time.UTC),
					tc.data,
				}

				output, err := log.Print(outputFormat)
				assert.Nil(t, err)
				assert.Equal(t, expectedOutput, output)
			})
		}
	}

	t.Run("Should return an error with an unknown output format", func(t *testing.T) {
		log := Log{
			LogLevelInfo,
			time.Date(1989, 6, 22, 7, 54, 0, 0, time.UTC),
			textMessage("this is a test log"),
		}

		_, err := log.Print(OutputFormat("eggcorn"))
		assert.Equal(t, errors.New("unsupported output format type: eggcorn"), err)
	})

	for _, tc := range []OutputFormat{OutputFormatText, OutputFormatJSON} {
		t.Run(fmt.Sprintf("Should propagate an error that occurs while producing %s output", tc), func(t *testing.T) {
			failLog := Log{LogLevelInfo, time.Now(), failMessage{}}
			_, err := failLog.Print(tc)
			assert.Equal(t, errFailMessage, err)
		})
	}
}

var errFailMessage = errors.New("something bad happened")

type failMessage struct{}

func (f failMessage) Message() (string, error) {
	return "", errFailMessage
}

func (f failMessage) Payload() ([]string, map[string]interface{}, error) {
	return nil, nil, errFailMessage
}
