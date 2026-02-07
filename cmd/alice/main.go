package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ===== RSS =====

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
}

// ===== Alice =====

type AliceRequest struct {
	Version string `json:"version"`
	Session struct {
		New bool `json:"new"`
	} `json:"session"`
	Request struct {
		Command string `json:"command"`
		Type    string `json:"type"`
	} `json:"request"`
}

type AliceResponse struct {
	Response struct {
		Text       string `json:"text"`
		TTS        string `json:"tts"`
		EndSession bool   `json:"end_session"`
	} `json:"response"`
	Version string `json:"version"`
}

const RSS_URL = "https://www.championat.com/rss/news/football"

func main() {
	http.HandleFunc("/webhook", aliceHandler)

	log.Println("Alice webhook started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func aliceHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	log.Println("REQUEST:", string(body))

	var req AliceRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	command := strings.ToLower(req.Request.Command)
	var text string

	switch {
	case req.Session.New:
		text = "Привет! Я могу рассказать последние футбольные новости. Скажи: новости футбола."

	case strings.Contains(command, "новост") && strings.Contains(command, "футбол"):
		text = getFootballNews()

	default:
		text = "Скажи: расскажи новости футбола."
	}

	resp := AliceResponse{Version: "1.0"}
	resp.Response.Text = text
	resp.Response.TTS = sanitizeForTTS(text)
	resp.Response.EndSession = false

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ===== News =====

func getFootballNews() string {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(RSS_URL)
	if err != nil {
		return "Не удалось получить футбольные новости."
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Ошибка чтения новостей."
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return "Ошибка обработки новостей."
	}

	limit := 10
	if len(rss.Channel.Items) < limit {
		limit = len(rss.Channel.Items)
	}

	result := "Вот последние новости футбола. "
	for i := 0; i < limit; i++ {
		result += fmt.Sprintf("%d. %s. ", i+1, rss.Channel.Items[i].Title)
	}

	return result
}

// Убираем мусор для голосовой озвучки
func sanitizeForTTS(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	return text
}
