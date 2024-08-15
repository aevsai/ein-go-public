package tools

import "testing"

func RestRequestInternetSearch(t *testing.T) {
	_, err := RequestIntenetSearch("golang")
    if err != nil {
        t.Fatalf("Error searching the internet: %s \n", err)
	}
}
