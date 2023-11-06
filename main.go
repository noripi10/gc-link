package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("sanko/gc-link")

	// eviroment variable
	env := getFilePath(".env", "")
	err := godotenv.Load(env)
	if err != nil {
		log.Fatalf("Unload .env")
	}
	sammary := os.Getenv("SUMMARY")

	ctx := context.Background()

	y, m, days := getParams()

	timeMin := getRfcTime(time.Date(y, m, 1, 0, 0, 0, 0, time.Local))
	getumatu := time.Date(y, m, 1, 23, 59, 59, 59, time.Local).AddDate(0, 1, -1)
	timeMax := getRfcTime(getumatu)

	// create http client
	client := getClient(ctx)
	// create calendar service
	service := getService(ctx, client)
	// 取得
	events, err := service.Events.List("primary").TimeMin(timeMin).TimeMax(timeMax).ShowDeleted(false).SingleEvents(true).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve events. %v", err)
	}
	log.Print(events)

	var createList []string = days
	var deleteList []string

	if len(events.Items) == 0 {
		for _, day := range createList {
			createEvent(*service, day, sammary)
		}
	} else {
		for _, item := range events.Items {
			if item.Summary != sammary {
				continue
			}

			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			date = strings.Replace(date, "-", "", -1)

			isExist := false
			for _, day := range days {
				if strings.Contains(day, date) {
					isExist = true

					break
				}
			}

			fmt.Printf("%v [%v]\n", item.Summary, date)

			if isExist {
				createList = removeTarget(createList, date)
				continue
			}
			// 削除
			if !isExist {
				deleteList = append(deleteList, item.Id)
				continue
			}
		}

		// TODO バッチ処理がライブラリに見当たらない...

		fmt.Println(createList)
		for _, day := range createList {
			createEvent(*service, day, sammary)
		}

		fmt.Println(deleteList)
		for _, eventId := range deleteList {
			deleteEvent(*service, eventId)
		}
	}
}
