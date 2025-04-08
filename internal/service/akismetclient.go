package service

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"tegobot/internal/logger"
	"time"
)

const ApiURL = "https://%s.rest.akismet.com/1.1/comment-check"

// Akismet структура для работы с Akismet API
type Akismet struct {
	apiKey     string
	blogURL    string
	httpClient *http.Client
	logger     *logger.Logger
}

type CommentData struct {
	Blog             string
	UserIp           string
	UserAgent        string
	CommentType      string
	CommentAuthor    string
	CommentContent   string
	CommentAuthorUrl string
	IsTest           bool
}

// NewAkismet создает новый клиент
func NewAkismet(apiKey, blogURL string, logger *logger.Logger) *Akismet {
	instance := new(Akismet)
	instance.apiKey = apiKey
	instance.blogURL = blogURL
	instance.logger = logger
	instance.httpClient = &http.Client{Timeout: 5 * time.Second}
	return instance
}

// MessageIsSpam CheckCommentIsSpam проверяет сообщение на спам через Akismet
func (ac *Akismet) MessageIsSpam(comment CommentData) (bool, error) {
	url := fmt.Sprintf(ApiURL, ac.apiKey)

	isTest := "false"
	if comment.IsTest {
		isTest = "true"
	}
	blogURL := ac.blogURL
	if comment.Blog != "" {
		blogURL = comment.Blog
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("blog", blogURL)
	_ = writer.WriteField("user_ip", comment.UserIp)
	_ = writer.WriteField("user_agent", comment.UserAgent)
	_ = writer.WriteField("comment_type", comment.CommentType)
	_ = writer.WriteField("comment_author", comment.CommentAuthor)
	_ = writer.WriteField("comment_content", comment.CommentContent)
	_ = writer.WriteField("comment_author_url", comment.CommentAuthorUrl)
	_ = writer.WriteField("is_test", isTest)

	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	ac.logger.WriteToJson("akismet CommentData", comment)

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("akismet API error: %s", resp.Status)
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	ac.logger.WriteToJson("akismet response", string(bodyText))
	var isSpam bool

	isSpam = string(bodyText) == "true"

	return isSpam, nil
}
