package processors

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/honeycombio/honeycomb-kubernetes-agent/event"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

var (
	ErrFilterOptionUnspecified = errors.New("drop_event processor requires a 'Field' to be set")
)

type EventDropper struct {
	config   *eventDropperConfig
	patterns []*regexp.Regexp
}

type eventDropperConfig struct {
	Field  string
	Values []string
}

func (f *EventDropper) Init(options map[string]interface{}) error {
	config := &eventDropperConfig{}
	err := mapstructure.Decode(options, config)
	if err != nil {
		return err
	}

	if config.Field == "" {
		return ErrFilterOptionUnspecified
	}
	f.config = config

	f.patterns = make([]*regexp.Regexp, len(f.config.Values))
	for i, val := range f.config.Values {
		compiled, _ := regexp.Compile("^" + val + "$")
		f.patterns[i] = compiled
	}

	return nil
}

func (f *EventDropper) Process(ev *event.Event) bool {
	// Keep event if no data
	if ev.Data == nil {
		return true
	}

	// Keep event if the event doesn't have this dropper's field
	val, ok := ev.Data[f.config.Field]
	if !ok {
		return true
	}

	// Convert value to string
	valString, ok := val.(string)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"key":   f.config.Field,
			"value": val,
			"type":  fmt.Sprintf("%T", val)}).
			Debug("Not filtering field of non-string type")
		return true
	}

	// Check if the value matches any of the provided patterns
	for _, re := range f.patterns {
		if re.MatchString(valString) {
			return false
		}
	}

	return true
}
