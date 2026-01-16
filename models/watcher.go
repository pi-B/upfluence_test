package models

import (
	"analysis-api/logging"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
)

type Watcher struct {
	TargetDimension string
	mu              sync.Mutex
	Report
}

type Report struct {
	FirstTimestamp   int
	LastTimestamp    int
	Counter          int
	TotalDimension   int
	AverageDimension int
}

var logger *slog.Logger

func NewWatcher(dimension string) Watcher {
	dimension = strings.Title(dimension) // Using this despite the deprecation to fit the stdlib limitation, so that the name of the dimension respects the case of the Go publics attributes
	return Watcher{TargetDimension: dimension}
}

// Make the starter start listening to the data containing the content we want to analyse
func (w *Watcher) Start(c chan map[string]SocialsData) {
	logger = logging.Get()
	for data := range c {
		var social SocialsData
		parsed_data := getSocialsData(data)
		if parsed_data == nil {
			continue
		}
		json.Unmarshal(parsed_data, &social)
		go w.addInput(social)
	}

}

func (w *Watcher) addInput(s SocialsData) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if s.Timestamp < w.Report.FirstTimestamp || w.Report.FirstTimestamp == 0 {
		w.Report.FirstTimestamp = s.Timestamp
	}
	if s.Timestamp > w.Report.LastTimestamp {
		w.Report.LastTimestamp = s.Timestamp
	}

	value := w.GetDimension(s)
	//logging dimension type and value for info
	w.TotalDimension += value
	w.Counter++
	logger.Debug(fmt.Sprintf("Results : %+v\n", w.Report))

}

func (w *Watcher) GetDimension(s SocialsData) int {
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		logger.Error("Kind is not reflect.Struct")
		return 0
	}

	field := v.FieldByName(w.TargetDimension)
	if !field.IsValid() {
		logger.Error("Field is not valid for " + w.TargetDimension)
		return 0
	}

	if !field.CanInterface() {
		logger.Error("Can't for interface" + w.TargetDimension)
		return 0
	}

	value := int(field.Int())
	return value
}

func (w *Watcher) Finalize() map[string]interface{} {

	w.AverageDimension = w.TotalDimension / w.Counter

	return map[string]interface{}{"minimum_timestamp": w.FirstTimestamp, "maximum_timestamp": w.LastTimestamp, "total_posts": w.Counter, "avg_" + w.TargetDimension: w.AverageDimension}
}

// Extract the data we want to handle from a JSON structured object if it falls into our scope
func getSocialsData(raw map[string]SocialsData) []byte {
	for _, social_type := range SOCIAL_TYPES {
		if d, ok := raw[social_type]; ok {
			data, err := json.Marshal(d)
			if err != nil {
				logger.Error("Could not marshal SocialData object",
					"error", err.Error())
				return nil
			}
			return data
		}
	}

	return nil
}
