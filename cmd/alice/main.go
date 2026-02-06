package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Структуры для парсинга RSS
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title string `xml:"title"`
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// Структура запроса Алисы
type AliceRequest struct {
	Version string `json:"version"`
	Request struct {
		Command string `json:"command"`
		Type    string `json:"type"`
		Intent  struct {
			Name string `json:"name"`
		} `json:"intent"`
	} `json:"request"`
}

// Структура ответа Алисы
type AliceResponse struct {
	Response struct {
		Text       string `json:"text"`
		EndSession bool   `json:"end_session"`
	} `json:"response"`
	Version string `json:"version"`
}

const RSS_URL = "https://www.championat.com/rss/news/football"

func main() {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		var req AliceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var responseText string

		switch req.Request.Type {
		case "LaunchRequest":
			responseText = "Привет! Я могу рассказать последние футбольные новости. Скажи 'новости футбола'."
		case "IntentRequest":
			switch req.Request.Intent.Name {
			case "GetFootballNews":
				responseText = getFootballNews()
			default:
				responseText = "Извини, я не понимаю этот запрос."
			}
		default:
			responseText = "Неизвестный тип запроса."
		}

		resp := AliceResponse{
			Version: "1.0",
		}
		resp.Response.Text = responseText
		resp.Response.EndSession = false

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getFootballNews() string {
	resp, err := http.Get(RSS_URL)
	if err != nil {
		return "Не удалось получить новости."
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Не удалось прочитать новости."
	}

	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return "Не удалось распарсить новости."
	}

	newsCount := 10
	if len(rss.Channel.Items) < newsCount {
		newsCount = len(rss.Channel.Items)
	}

	newsText := "Вот последние новости футбола:\n"
	for i := 0; i < newsCount; i++ {
		item := rss.Channel.Items[i]
		// Сокращаем описание для озвучки
		newsText += fmt.Sprintf("%d. %s. %s\n", i+1, item.Title, item.Description)
	}

	return newsText
}
