// Package tools
// Author: Evsikov Artem

package tools

import (
	"context"
	"fmt"
	"time"

	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

const (
	apiKey = "AIzaSyBaicxfw17BG6PLwtoRbbnoVmt2frNMG0k"
	cx     = "86f0d151b228a40f6"
)

func RequestIntenetSearch(query string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	service, err := customsearch.NewService(ctx, option.WithAPIKey(apiKey))

	if err != nil {
		logger.Err(err).Msg("")
        return "Error while searching in the internet.", err
	}

	resp, err := service.Cse.List().Cx(cx).Q(query).Do()

	if resp.HTTPStatusCode != 200 {
        logger.Error().Msgf("Unsuccessful api code: %d", resp.HTTPStatusCode)
		return "Error while searching in the internet.", err
	}

	for _, v := range resp.Items {
		return fmt.Sprintf("Title: %s; Link: %s; Description: %s", v.HtmlTitle, v.DisplayLink, v.HtmlSnippet), nil
	}

	return "Empty result.", nil
}
