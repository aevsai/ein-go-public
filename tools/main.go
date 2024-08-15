// Package tools
// Author: Evsikov Artem

package tools

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog"
)

var logger = zerolog.New(os.Stdout)

var AvailableTools = []string{"insert_event_into_calendar", "add_new_calendar", "get_calendar_events", "get_current_weather", "search_internet", "read_file"}

func GetFunctionParamsByName(name string) map[string]interface{} {
	config, err := os.ReadFile("/app/tools/configs/" + name + ".json")
	if err != nil {
		logger.Err(err).Msg("")
    }
	var objMap map[string]interface{}
	err = json.Unmarshal(config, &objMap)
	if err != nil {
		logger.Err(err).Msg("")
    }
	return objMap
}

func GetFunctionByName(name string) func(string) (string, error) {
	var mappedFunctions = make(map[string]func(string) (string, error))

	//Declare available functions for function calling
	mappedFunctions["get_current_weather"] = RequestWeather
	mappedFunctions["search_internet"] = RequestIntenetSearch
    mappedFunctions["read_file"] = SummarizeFile
    mappedFunctions["get_calendar_events"] = GetEvent
    mappedFunctions["add_new_calendar"] = AddCalendar
    mappedFunctions["insert_event_into_calendar"] = CreateEvent
	return mappedFunctions[name]
}
