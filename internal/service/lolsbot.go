package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// see https://lols.bot/

type LolsBot struct {
	httpClient *http.Client
}

type ResponseLolsBot struct {
	Ok         bool    `json:"ok"`
	UserId     int64   `json:"user_id"`
	Banned     bool    `json:"banned"`
	When       string  `json:"when"`
	Offenses   int     `json:"offenses"`
	SpamFactor float64 `json:"spam_factor"`
}

func NewLolsBot() *LolsBot {
	instance := new(LolsBot)
	instance.httpClient = &http.Client{Timeout: 5 * time.Second}
	return instance
}

func (l *LolsBot) Verify(userID int64) (bool, error) {
	var err error
	url := fmt.Sprintf("https://api.lols.bot/account?id=%d", userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:136.0) Gecko/20100101 Firefox/136.0")
	req.Header.Add("Cookie", fmt.Sprintf("x-user-ip=%s", GeneratePublicIPv4()))
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Host", "api.lols.bot")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	response := new(ResponseLolsBot)
	err = json.Unmarshal(bodyBytes, response)
	if err != nil {
		return false, err
	}

	return response.Banned, nil

}

func (l *LolsBot) GetServiceName() string {
	return "LolsBot"
}
