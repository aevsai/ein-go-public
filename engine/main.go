package engine

import (
	"einstein-server/chatgpt"
	"einstein-server/database"
	"einstein-server/tools"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var logger = zerolog.New(os.Stdout)

func RunEngine() {
    s, err := gocron.NewScheduler()
	if err != nil {
		logger.Err(err).Msg("")
        return
	}

	// add a job to the scheduler
	j, err := s.NewJob(
		gocron.DurationJob(60*time.Second),
		gocron.NewTask(runFlow),
	)
	if err != nil {
		logger.Err(err).Msg("")
        return
	}
	// each job has a unique id
	fmt.Println(j.ID())

	// start the scheduler
	s.Start()
    defer s.Shutdown()
}

// calendarFlow requests data from calendar and complete flow
// 1. Request data
// 2. Get possible user intents based on data
// 3. Fill schema
// 4. What intents can be accomplished by bot
// 5. Create drafts and ask for confiramation
// 6. Complete confirmed tasks
// 7. Add other tasks to ToDo list
func runFlow() {
    db := database.GetConnection()
    var users []database.User
    if err := db.Select(&users, database.SqlUserSelectAll); err != nil {
        logger.Err(err).Msg("")
    }
    for _, user := range users {
        data, err := getCalendarContext(user.ID.UUID)
        if err != nil {
            logger.Err(err).Msg("")
            continue
        }
        var messages []interface{}
        messages = append(messages, chatgpt.ChatMessage{
            Role: database.RoleUser,
            Content: fmt.Sprintf("What intents user may have after receiving this messages.\n %s", data),
        })
        resp, err := chatgpt.ChatRequest(messages, nil, user.ID.UUID)
        if err != nil {
            logger.Err(err).Msg("")
            continue
        }
        messages = append(messages, chatgpt.ChatMessage{
            Role: database.RoleAssistant,
            Content: resp.Choices[0].Message.Content,
        })
        messages = append(messages, chatgpt.ChatMessage{
            Role: database.RoleUser,
            Content: `Rewrite intentions into json. By following schema
            {
                "type": "object",
                "params": {
                    "description": {
                        "type": "string",
                        "description": "Parameter that describes user intention"
                    },
                    "can_accomplish": {
                        "type": "bool",
                        "description": "true if assistant can accomplish this task, else false"
                    }
                }
            }
            `,
        })
        resp, err = chatgpt.ChatRequest(messages, nil, user.ID.UUID)
        if err != nil {
            logger.Err(err).Msg("")
            continue
        }
        s := resp.Choices[0].Message.Content
        start := strings.Index(s, "{")
        end := strings.LastIndex(s, "}")
        jsonString := s[start:end]
        var args map[string]string
        if err = json.Unmarshal([]byte(jsonString), &args); err != nil {
            logger.Err(err).Msg("")
            continue
        }
        logger.Debug().Msg(jsonString)
    }
}

func getEmailContext() (string, error) {
	return "", nil
}

func getCalendarContext(userID uuid.UUID) (string, error) {
    var input map[string]string 
    input["user_id"] = userID.String()
    input["events_limit"] = "20"
    input["time_from"] = time.Now().Format(time.RFC3339)
    input["time_to"] = time.Now().Add(24 * time.Hour).Format(time.RFC3339)
    v, err := json.Marshal(input)
    if err != nil {
       return "", err 
    }
    return tools.GetEvent(string(v))
}
