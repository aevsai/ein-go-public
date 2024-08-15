package tools

import "testing"

func TestGetFunctionParamsByName(t *testing.T) {
	data := GetFunctionParamsByName("get_current_weather")
	if data == nil {
		t.Fatalf("Error while reading file: %+v", data)
	}
}
