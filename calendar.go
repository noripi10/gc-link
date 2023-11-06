package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getTokenFromWeb(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("ブラウザでこのURLを開いて認証コードを取得して下さい\n%v\n", authURL)
	fmt.Print("取得した認証コードをコピーしてEnterを押して下さい\n")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return token
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	if err != nil {
		return nil, err
	}

	// validation
	// refresh tokenは利用頻度が少ないから無効化される可能性があるので
	// access tokenが無効化されたら再認証を入れる
	client := &http.Client{}
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?access_token=%v", token.AccessToken)
	res, err := client.Get(url)

	if err != nil {
		return nil, errors.New("http request error")
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)

	_, ok := data["expires_in"]

	if !ok {
		return nil, errors.New("access token invalided")
	}

	return token, nil
}

func getClient(ctx context.Context) *http.Client {
	// jsonData := `{}`
	// var data map[string]interface{}
	// if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
	// 	log.Fatalf("Failed to parse JSON data: %v", err)
	// }

	fullPath := getFilePath("oauth_client.json", "")
	bs, err := os.ReadFile(fullPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(bs, calendar.CalendarScope)
	config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
	if err != nil {
		log.Fatal(err)
	}

	tokenFilePath := getFilePath("token.json", "")
	token, err := tokenFromFile(tokenFilePath)
	if err != nil {
		// アクセストークンが切れたら削除
		os.Remove(tokenFilePath)
		token = getTokenFromWeb(ctx, config)
		saveToken(tokenFilePath, token)
	}
	return config.Client(ctx, token)
}

func getService(ctx context.Context, client *http.Client) *calendar.Service {
	svc, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Service create error %v", err)
	}
	return svc
}

func getRfcTime(t time.Time) string {
	utcTime := t.UTC()
	timeString := utcTime.Format(time.RFC3339)

	return timeString
}

func createEvent(servie calendar.Service, day string, summary string) (*calendar.Event, error) {
	formatDay := Substr(day, 0, 4) + "-" + Substr(day, 4, 6) + "-" + Substr(day, 6, 8)

	event := &calendar.Event{
		Summary: summary,
		Start: &calendar.EventDateTime{
			Date:     formatDay,
			TimeZone: "Asia/Tokyo",
		},
		End: &calendar.EventDateTime{
			Date:     formatDay,
			TimeZone: "Asia/Tokyo",
		},
	}
	event, err := servie.Events.Insert("primary", event).Do()
	return event, err
}

func deleteEvent(service calendar.Service, eventId string) error {
	err := service.Events.Delete("primary", eventId).Do()
	return err
}
