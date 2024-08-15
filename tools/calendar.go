package tools

import (
	"bytes"
	"context"
	"database/sql"
	"einstein-server/database"
	"einstein-server/oauth"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, userID uuid.UUID, email ...string) ([]*http.Client, error) {
	db := database.GetConnection()
	defer db.Close()
	var users []database.User
	if err := db.Select(&users, database.SqlUserSelectByParentId, userID); err != nil {
		if err == sql.ErrNoRows {
			users = []database.User{}
		} else {
			logger.Err(err).Msg("")
			return nil, err
		}
	}
	var tokens []*oauth2.Token
	tok, err := oauth.GetUserIdPToken(userID, "google-oauth2")
	tokens = append(tokens, &tok)
	for _, user := range users {
		if len(email) == 0 {
			tok, err := oauth.GetUserIdPToken(user.ID.UUID, "google-oauth2")
			if err != nil {
				logger.Err(err).Msg("")
				continue
			}
			tokens = append(tokens, &tok)
		} else if user.Email == email[0] {
			tok, err := oauth.GetUserIdPToken(user.ID.UUID, "google-oauth2")
			if err != nil {
				logger.Err(err).Msg("")
				continue
			}
			tokens = append(tokens, &tok)
		}
	}
	logger.Debug().Msgf("%+v", tokens)
	if len(tokens) < 1 {
		if err = requestAuthorization(userID); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Authorization requested")
	}
	var clients []*http.Client
	for _, tok := range tokens {
		clients = append(clients, config.Client(context.Background(), tok))
	}
	return clients, nil
}

func requestAuthorization(userID uuid.UUID) error {
	db := database.GetConnection()
	defer db.Close()
	var user database.User
	if err := db.Get(&user, database.SqlUserSelect, userID); err != nil {
		logger.Err(err).Msg("")
		return err
	}

	b, err := json.Marshal([]database.SmartContent{
		{
			Type:  database.SmartContentTypeButton,
			Key:   database.SmartContentKeyGoogleAuth,
			Value: "",
		},
	})
	var jsonSmartContent json.RawMessage
	jsonSmartContent.UnmarshalJSON(b)
	if err != nil {
		logger.Err(err).Msg("")
		return err
	}
	message := database.Message{
		ID:           uuid.New(),
		UserId:       userID,
		Content:      "Authorize via Google to use calendar.",
		SmartContent: &jsonSmartContent,
		Role:         database.RoleAssistant,
	}
	if _, err := db.NamedExec(database.SqlMessageInsert, message); err != nil {
		logger.Err(err).Msg("")
		return err
	}

	return nil
}

func GetEvent(arguments string) (string, error) {
	var args map[string]interface{}
	json.NewDecoder(bytes.NewReader([]byte(arguments))).Decode(&args)

	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		logger.Err(err).Msg("")
		return "Error while requesting events", err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		logger.Err(err).Msg("")
		return "Error while requesting events", err
	}
	var clients []*http.Client
	for i := 1; i <= 3; i++ {
		clients, err = getClient(config, uuid.MustParse(args["user_id"].(string)))
		if err != nil {
			if err.Error() == "Requested authorization" {
				time.Sleep(60 * time.Second)
			} else {
				return "Error while requesting events", err
			}
			logger.Err(err).Msg("")
		} else {
			break
		}
	}
	var allEvents string
	for _, client := range clients {
		srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			logger.Err(err).Msg("")
			return "Error while requesting events", err
		}
		eventSrv := srv.Events.
			List("primary").
			ShowDeleted(false).
			SingleEvents(true).
			OrderBy("startTime")

		if args["events_limit"] != nil {
			eventSrv = eventSrv.MaxResults(int64(args["events_limit"].(float64)))
		}
		var minTime time.Time
		var maxTime time.Time
		if args["time_from"] != nil {
			if err = minTime.UnmarshalText([]byte(args["time_from"].(string))); err == nil {
				eventSrv = eventSrv.TimeMin(minTime.Format(time.RFC3339))
			}
		}
		if args["time_to"] != nil {
			if err = maxTime.UnmarshalText([]byte(args["time_to"].(string))); err == nil {
				eventSrv = eventSrv.TimeMax(maxTime.Format(time.RFC3339))
			}
		}
		events, err := eventSrv.Do()
		if err != nil {
			logger.Err(err).Msg("")
			continue
		} else {
			bEvents, err := events.MarshalJSON()
			if err != nil {
				logger.Err(err).Msg("")
				continue
			}
			allEvents = allEvents + string(bEvents)
		}
	}
	if len(allEvents) > 0 {
		return allEvents, nil
	} else {
		return "User have no events", nil
	}
}

func AddCalendar(arguments string) (string, error) {
	var args map[string]interface{}
	json.NewDecoder(bytes.NewReader([]byte(arguments))).Decode(&args)
	if err := requestAuthorization(uuid.MustParse(args["user_id"].(string))); err != nil {
		return "Error requesting authorization", err
	}
	return "Authorization instructions have been sent to user. Say him that hw should click the button in the message above and choose account he want to connect", nil
}

func CreateEvent(arguments string) (string, error) {
	var args map[string]interface{}
	json.NewDecoder(bytes.NewReader([]byte(arguments))).Decode(&args)
	if args["user_id"] == nil || args["email"] == nil || args["start_at"] == nil || args["summary"] == nil {
		logger.Err(fmt.Errorf("One of required arguments is nil: %+v", args)).Msg("")
		return "One of required arguments is nil:", fmt.Errorf("One of required arguments is nil: %+v", args)
	}
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		logger.Err(err).Msg("")
		return "Error while requesting events", err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		logger.Err(err).Msg("")
		return "Error while requesting events", err
	}
	var clients []*http.Client
	clients, err = getClient(config, uuid.MustParse(args["user_id"].(string)), args["email"].(string))
	if len(clients) < 1 {
		logger.Err(err).Msg("")
		return "Error while requesting events", err
	}
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(clients[0]))
	if err != nil {
		logger.Err(err).Msg("")
		return "Error while requesting events", err
	}

	event := &calendar.Event{
		Summary: args["summary"].(string),
	}

	if args["location"] != nil {
		event.Location = args["location"].(string)
	}
	if args["description"] != nil {
		event.Description = args["description"].(string)
	}
	if args["start_at"] != nil {
		event.Start = &calendar.EventDateTime{
			DateTime: args["start_at"].(string),
			TimeZone: "Etc/UTC",
		}
	}
	if args["end_at"] != nil {
		event.End = &calendar.EventDateTime{
			DateTime: args["end_at"].(string),
			TimeZone: "Etc/UTC",
		}
	}
	if args["attendees"] != nil {
		attendees := strings.Split(args["attendees"].(string), ",")
		for _, a := range attendees {
			event.Attendees = append(event.Attendees, &calendar.EventAttendee{Email: a})
		}
	}

	calendarId := "primary"
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		logger.Err(err).Msg("Unable to create event.")
		return "Unable to create event.", err
	}
	return fmt.Sprintf("Event created: %s\n", event.HtmlLink), nil
}
