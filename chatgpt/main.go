// Package chatgpt
// Author: Evsikov Artem

package chatgpt

import (
	"bytes"
	"database/sql"
	"einstein-server/database"
	"einstein-server/tools"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

const ApiToken = ""
const SummarizePrompt = "Summarize following text: \n %s"

var logger = zerolog.New(os.Stdout).Level(zerolog.DebugLevel)

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message      ChatMessageWithToolRequest `json:"message"`
		FinishReason string                     `json:"finish_reason"`
		Index        int                        `json:"index"`
	} `json:"choices"`
}

type ToolCalls struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatMessageWithToolResponse struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	ToolCallId string `json:"tool_call_id"`
}

type ChatMessageWithToolRequest struct {
	Role      string      `json:"role"`
	Content   string      `json:"content"`
	ToolCalls []ToolCalls `json:"tool_calls"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type FunctionParameter struct {
	Type       string      `json:"type"`
	Properties interface{} `json:"properties"`
	Required   []string    `json:"required"`
}

type Function struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  FunctionParameter `json:"parameters"`
}

type ChatCompletionRequest struct {
	Model       string                   `json:"model"`
	Messages    []interface{}            `json:"messages"`
	Tools       []map[string]interface{} `json:"tools"`
	ToolChoice  string                   `json:"tool_choice,omitempty"`
	Temperature float32                  `json:"temperature"`
}

type ErrorString struct {
	s string
}

func (e *ErrorString) Error() string {
	return e.s
}

func NewError(text string) error {
	return &ErrorString{text}
}

func SaveUsageStatistics(amountIn int, amountOut int, messagesAmount int, userId uuid.UUID) {
	var us database.UsageStatistics

	db := database.GetConnection()
	defer db.Close()
	err := db.Get(&us, database.SqlUsageStatisticsSelectByUserAndDate, userId, time.Now())
	if err != nil {
		if err == sql.ErrNoRows {
			us = database.UsageStatistics{ID: uuid.New(), UserId: userId, Date: time.Now(), TokensIn: 0, TokensOut: 0}
			db.NamedExec(database.SqlUsageStatisticsInsert, us)
		} else {
			logger.Error().Msgf("Error while selecting usage_statistics: %s \n", err)
		}
	}
	us.TokensIn += amountIn
	us.TokensOut += amountOut
	us.MessagesAmount += messagesAmount
	db.NamedExec(database.SqlUsageStatisticsUpdate, us)
	db.Close()
}

/*
Make context that follow conditions
- If inContextForce = true, then mandatory include to context
- If inContext = false, then exclude from context
- If context capacity exceeded, then include mandatory messages first. Clip messages by timestamp (clip the oldest ones) to fit context.
*/
func ClipContentToContextWindow(messages []database.Message) []database.Message {
	var totalContent string
	var contextMessages []database.Message

	for i := len(messages) - 1; i >= 0; i-- {
		if !messages[i].InContextByForce {
			continue
		}
		contextMessages = append(contextMessages, messages[i])
		if len(messages[i].SummarizedContent.String) > 0 {
			totalContent += messages[i].SummarizedContent.String
		} else {
			totalContent += messages[i].Content
		}
	}

	for i := len(messages) - 1; i >= 0; i-- {
		if !messages[i].InContext {
			continue
		}
		contextMessages = append(contextMessages, messages[i])
		if len(messages[i].SummarizedContent.String) > 0 {
			totalContent += messages[i].SummarizedContent.String
		} else {
			totalContent += messages[i].Content
		}
		// for worst case where 1 symbol = 1 token make filter + leave something on tools and technical info (e.g. file ids)
		if len(totalContent) > 10000 {
			break
		}
	}
	sort.SliceStable(contextMessages, func(i, j int) bool {
		return contextMessages[i].CreatedAt.Before(contextMessages[j].CreatedAt)
	})
	return contextMessages
}

func ChatRequest(messages []interface{}, tools []map[string]interface{}, userId uuid.UUID) (ChatCompletionResponse, error) {
	const url = "https://api.openai.com/v1/chat/completions"
	var toolChoice string
	if len(tools) > 0 {
		toolChoice = "auto"
	}
	jsonBody := ChatCompletionRequest{
		Model:       "gpt-3.5-turbo-16k",
		Messages:    messages,
		Tools:       tools,
		ToolChoice:  toolChoice,
		Temperature: 0.4,
	}

	b, err := json.Marshal(jsonBody)
	logger.Debug().Msg(string(b))
	if err != nil {
		logger.Err(err)
		return ChatCompletionResponse{}, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		logger.Err(err)
		return ChatCompletionResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ApiToken))

	client := http.Client{
		Timeout: 180 * time.Second,
	}

	res, err := client.Do(req)

	if err != nil {
		logger.Err(err)
		return ChatCompletionResponse{}, err
	}
	var data ChatCompletionResponse

	if res.StatusCode != 200 {
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, res.Body); err != nil {
			buf = bytes.NewBuffer([]byte("Error while reading body"))
		}
		logger.Printf("%d %+v", res.StatusCode, buf.String())

		return ChatCompletionResponse{}, fmt.Errorf("Unseccessful status code: %d", res.StatusCode)
	}
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return ChatCompletionResponse{}, err
	}
	SaveUsageStatistics(data.Usage.PromptTokens, data.Usage.CompletionTokens, 0, userId)
	return data, nil
}

func InjectDataIntoMessages(clippedMessages []database.Message) []interface{} {
	var content string
	var messagesGPT []interface{}
	for _, v := range clippedMessages {
		if len(v.SummarizedContent.String) > 0 {
			content = v.SummarizedContent.String
		} else {
			content = v.Content
		}
		if v.Attachments != nil && len(v.Attachments) > 0 {
			for _, f := range v.Attachments {
				content = fmt.Sprintf("My file has id: %s \n", f.ID) + content
			}
		}
		if v.Role == database.RoleUser {
			content = fmt.Sprintf("My user id: %s \n", v.UserId) + content
			content = fmt.Sprintf("Message has been sent at: %s \n", v.CreatedAt.Format(time.RFC1123)) + content
		}
		messagesGPT = append(messagesGPT, ChatMessage{Role: v.Role, Content: content})
	}
	return messagesGPT
}

func RequestCompletion(user database.User, db *sqlx.DB, recursionMessages *[]interface{}) (ChatCompletionResponse, error) {
	var err error
	var response ChatCompletionResponse
	var messagesGPT []interface{}

	// Prepare messages
	if recursionMessages == nil {
		messages, err := database.SelectMessageByUser(user)
		if err != nil {
			return response, err
		}

		clippedMessages := ClipContentToContextWindow(messages)
		if len(clippedMessages) == 0 {
			return ChatCompletionResponse{}, NewError("Message too long. Please shorten it. 😅")
		}

		messagesGPT = InjectDataIntoMessages(clippedMessages)
	} else {
		messagesGPT = *recursionMessages
	}
	// Prepare tools
	var availableToolsGPT []map[string]interface{}
	for _, v := range tools.AvailableTools {
		availableToolsGPT = append(availableToolsGPT, tools.GetFunctionParamsByName(v))
	}

	// Make request
	response, err = ChatRequest(messagesGPT, availableToolsGPT, user.ID.UUID)
	if err != nil {
		return response, err
	}
	if len(response.Choices) == 0 {
		return response, fmt.Errorf("Zero choices returned")
	}

	// Condition for recursion
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		messagesGPT = append(messagesGPT, response.Choices[0].Message)
		for _, call := range response.Choices[0].Message.ToolCalls {
			callable := tools.GetFunctionByName(call.Function.Name)
			logger.Printf("Calling tool with name: %s; params: %s;\n", call.Function.Name, call.Function.Arguments)
			result, err := callable(call.Function.Arguments)
			if err != nil {
				logger.Err(err).Msg("")
			}
			messagesGPT = append(messagesGPT, ChatMessageWithToolResponse{
				ToolCallId: call.ID,
				Role:       database.RoleTool,
				Content:    result,
			})
		}
		response, err = RequestCompletion(user, db, &messagesGPT)
		if err != nil {
			return response, err
		}
		if len(response.Choices) == 0 {
			return response, fmt.Errorf("Error requesting chatgpt: %+v", response)
		}
	} else {
		answer := database.Message{
			ID:      uuid.New(),
			UserId:  user.ID.UUID,
			Content: response.Choices[0].Message.Content,
			Role:    database.RoleAssistant,
		}
		if len(answer.Content) > 200 {
			var sumPrompt []interface{}
			sumPrompt = append(sumPrompt, ChatMessage{
				Content: fmt.Sprintf(SummarizePrompt, answer.Content),
				Role:    database.RoleUser,
			})
			sum, _ := ChatRequest(sumPrompt, nil, user.ID.UUID)
			if len(sum.Choices) > 0 {
				answer.SummarizedContent = sql.NullString{String: sum.Choices[0].Message.Content, Valid: true}
			}
		}
		_, err = db.NamedExec(database.SqlMessageInsert, answer)
		if err != nil {
			return response, err
		}

		logger.Debug().Msgf("Got new message %+v \n", response)
		SaveUsageStatistics(0, 0, 1, user.ID.UUID)
	}
	return response, nil
}
