package service

import (
	"com-service/internal/models"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Read data from csv file given
func ReadCSVFile(filePath string) models.CustomerSchedules {
	log.Printf("stated reading file %s\n", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("unable to locate the file: "+filePath, err)
	}

	defer file.Close()

	csvReader := csv.NewReader(file)
	data, err := csvReader.ReadAll()

	if err != nil {
		log.Fatal("issue in reading file "+filePath, err)
	}

	customerList := make([]models.Customer, len(data)-1)

	for i, r := range data[1:] {
		customerList[i] = models.Customer{Email: r[0], Text: r[1], Schedule: r[2]}
	}

	log.Println("done with reading file....!")
	log.Printf("%v\n", customerList)

	return models.CustomerSchedules{
		List: customerList,
	}
}

func ComposeSchedules(cutomerSchedules models.CustomerSchedules) []*models.CustomerEvent {
	fmt.Println("")
	log.Println("started marking schedules")
	customerEvents := make([]*models.CustomerEvent, len(cutomerSchedules.List))
	for i, c := range cutomerSchedules.List {
		events := getEvents(c.Schedule)
		customerEvents[i] = &models.CustomerEvent{
			Email:   c.Email,
			Message: c.Text,
			Closed:  false,
			Events:  events,
		}
	}
	log.Println("done with scheduling...!")

	return customerEvents
}

func getEvents(schedulePatter string) []*models.Event {
	slots := strings.Split(schedulePatter, "-")

	if len(slots) == 0 {
		return []*models.Event{}
	}

	events := make([]*models.Event, len(slots))
	for i, s := range slots {
		delay, unit := parseTimePattern(s)
		when := time.Now()

		switch strings.ToLower(unit) {
		case "s":
			when = when.Add(time.Second * time.Duration(delay))
		case "m":
			when = when.Add(time.Minute * time.Duration(delay))
		case "h":
			when = when.Add(time.Hour * time.Duration(delay))
		}

		events[i] = &models.Event{
			When: when,
		}
	}

	return events
}

func parseTimePattern(pattern string) (delay int, unit string) {

	var l, n []rune
	for _, r := range pattern {
		switch {
		case r >= 'A' && r <= 'Z':
			l = append(l, r)
		case r >= 'a' && r <= 'z':
			l = append(l, r)
		case r >= '0' && r <= '9':
			n = append(n, r)
		}
	}

	num, err := strconv.Atoi(string(n))
	if err != nil {
		log.Println("issue in parsing schedule pattern ", err)
	}

	return num, string(l)
}
