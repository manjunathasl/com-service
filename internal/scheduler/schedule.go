package scheduler

import (
	"bytes"
	"com-service/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Message struct {
	Email string `json:"email"`
	Text  string `json:"text"`
}

type ComResponse struct {
	Email string `json:"email"`
	Text  string `json:"text"`
	Paid  bool   `json:"paid"`
}

type Scheduler struct {
	Customers []*models.CustomerEvent
	ComURL    string
}

func New(customers []*models.CustomerEvent, comUrl string) Scheduler {
	return Scheduler{
		Customers: customers,
		ComURL:    comUrl,
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(1 * time.Second)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		log.Println("Started scheduler process communications")
		done := make(chan struct{})
		s.getEventsToProcess(time.Now(), done)
		for {
			select {
			case <-ticker.C:
				s.getEventsToProcess(time.Now(), done)
			case <-done:
				ticker.Stop()
				for _, c := range s.Customers {
					log.Printf("%s %t\n", c.Email, c.Closed)
				}

				wg.Done()
				return
			}
		}
	}()
	wg.Wait()
}

func (s *Scheduler) postMessage(message Message) (ComResponse, error) {
	var comResp ComResponse
	jsonMessage, err := json.Marshal(message)

	if err != nil {
		return comResp, fmt.Errorf("error in encoding message %v", err)
	}

	reqBody := bytes.NewBuffer(jsonMessage)
	res, err := http.Post(s.ComURL, "application/json", reqBody)
	if err != nil {
		return comResp, fmt.Errorf("error in com service invoke %v", err)
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&comResp)

	if res.StatusCode != 201 {
		return comResp, errors.New("could not post the message succesfully")
	}

	return comResp, nil
}

func (s *Scheduler) getEventsToProcess(now time.Time, done chan struct{}) {

	closedSchedulesCount := 0
	for _, c := range s.Customers {
		if !c.Closed {
			eventConter := 0
			for _, e := range c.Events {
				if !c.IsProcessing() && e.NotifiedAt.IsZero() && (e.When.Before(now) || e.When == now) {

					go func(c *models.CustomerEvent, e *models.Event) {
						if c.IsProcessing() {
							return
						}
						c.Start()
						log.Printf("notifying User: %s", c.Email)

						data, err := s.postMessage(Message{
							Email: c.Email,
							Text:  c.Message,
						})

						if err != nil {
							log.Printf("unable to communicate com service %s \n", c.Email)
							return
						}

						if data.Paid {
							c.Done()
						}
						e.NotifiedAt = time.Now()
						c.Stop()
						log.Printf("notified '%s' at %s paid: %t\n", data.Email, e.NotifiedAt.Format(time.ANSIC), data.Paid)
					}(c, e)
				} else if !e.NotifiedAt.IsZero() {
					eventConter += 1
				}
			}
			if eventConter == len(c.Events) {
				c.Done()
			}

		} else {
			closedSchedulesCount += 1
		}
	}

	if closedSchedulesCount == len(s.Customers) {
		close(done)
	}
}
