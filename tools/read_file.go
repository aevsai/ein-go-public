package tools

import (
	"bytes"
	"database/sql"
	"einstein-server/database"
	"einstein-server/storage"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

const ApiToken = ""

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
		Message      ChatMessage    `json:"message"`
		FinishReason string         `json:"finish_reason"`
		Index        int            `json:"index"`
	} `json:"choices"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model       string                   `json:"model"`
	Messages    []ChatMessage            `json:"messages"`
	Temperature float32                  `json:"temperature"`
}

func SaveUsageStatistics(amountIn int, amountOut int, userId uuid.UUID) {
	var us database.UsageStatistics

	db := database.GetConnection()
    defer db.Close()
	err := db.Get(&us, database.SqlUsageStatisticsSelectByUserAndDate, userId, time.Now())
	if err != nil {
		if err == sql.ErrNoRows {
			us = database.UsageStatistics{ID: uuid.New(), UserId: userId, Date: time.Now(), TokensIn: 0, TokensOut: 0}
			db.NamedExec(database.SqlUsageStatisticsInsert, us)
		}
		logger.Printf("Error while selecting usage_statistics: %s \n", err)
	}
	us.TokensIn += amountIn
	us.TokensOut += amountOut
	db.NamedExec(database.SqlUsageStatisticsUpdate, us)
}


func ChatRequest(messages []ChatMessage, userId uuid.UUID) (ChatCompletionResponse, error) {
	const url = "https://api.openai.com/v1/chat/completions"

	jsonBody := ChatCompletionRequest{
		Model:       "gpt-3.5-turbo",
		Messages:    messages,
		Temperature: 0.7,
	}

	b, err := json.Marshal(jsonBody)
	if err != nil {
        logger.Err(err).Msg("")
        return ChatCompletionResponse{}, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		logger.Err(err).Msg("")
        return ChatCompletionResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ApiToken))

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(req)

	if err != nil {
		logger.Err(err).Msg("")
        return ChatCompletionResponse{}, err
	}
	var data ChatCompletionResponse

	logger.Printf("%d %+v", res.StatusCode, res.Body)

	if res.StatusCode != 200 {
		logger.Err(err).Msg("")
        return ChatCompletionResponse{}, fmt.Errorf("Unseccessful status code: %d", res.StatusCode)
	}
	err = json.NewDecoder(res.Body).Decode(&data)
    if err != nil {
        return ChatCompletionResponse{}, err
    }
	SaveUsageStatistics(data.Usage.PromptTokens, data.Usage.CompletionTokens, userId)
	return data, nil
}

func readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	// remember close file
    defer f.Close()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
    b, err := r.GetPlainText()
    if err != nil {
        return "", err
    }
    buf.ReadFrom(b)
	return buf.String(), nil
}

func SummarizeFile(arguments string) (string, error) {
    var parsedArguments map[string]string
    err := json.NewDecoder(bytes.NewReader([]byte(arguments))).Decode(&parsedArguments)
    if err != nil {
        return "Error while reading file.", err
    }
    db := database.GetConnection()
    defer db.Close()
    var attachment database.Attachment
    err = db.Get(&attachment, database.SqlAttachmentSelectById, parsedArguments["file_id"])
    if err != nil {
        return "Error while reading file.", err
    }
    strg := storage.NewClient()
    name, err := strg.GetFile(attachment.Key)
    if err != nil {
        return "Error while reading file.", err
    }

    defer os.Remove(name)
    content, err := readPdf(name)
    if err != nil {
        return "Error while reading file.", err
    }
    db.Close()
    return content, nil
//    messages := []ChatMessage{{Role: database.RoleUser, Content: fmt.Sprintf("Summarize following text: \n %s", content)}}
//    res, err := ChatRequest(messages, uuid.MustParse(parsedArguments["user_id"]))
//    if err != nil {
//        return "Error while summarizing file.", err
//    }
//    logger.Printf("Response for summarization: %+v", res)
//    return res.Choices[0].Message.Content, nil
}
