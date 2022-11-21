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
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
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
		fmt.Println("")
		log.Println("scheduler started...!")
		done := make(chan struct{})
		s.getEventsToProcess(time.Now(), done)
		for {
			select {
			case <-ticker.C:
				s.getEventsToProcess(time.Now(), done)
			case <-done:
				ticker.Stop()
				fmt.Println("")
				fmt.Println("      Done with sending notifications        ")
				for _, c := range s.Customers {
					log.Printf("%s Done: %t\n", c.Email, c.Closed)
				}
				processes, _ := process.Processes()
				for _, process := range processes {
					name, _ := process.Name()
					if name == "commservice.linux" {
						process.SendSignal(syscall.SIGINT)
					}
				}
				<-time.After(3 * time.Second)
				wg.Done()
				return
			}
		}
	}()
	wg.Wait()
}

func (s *Scheduler) postMessage(message Message) (ComResponse, error) {
	var comResp ComResponse

	client := http.Client{
		Timeout: 10 * time.Second,
	}
	jsonMessage, err := json.Marshal(message)

	if err != nil {
		return comResp, fmt.Errorf("error in encoding message %v", err)
	}

	reqBody := bytes.NewBuffer(jsonMessage)
	res, err := client.Post(s.ComURL, "application/json", reqBody)
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
					go s.checkAndCommunicate(c, e, done)
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

func (s *Scheduler) checkAndCommunicate(c *models.CustomerEvent, e *models.Event, done chan struct{}) {
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
		close(done)
		return
	}

	if data.Paid {
		c.Done()
	}
	e.NotifiedAt = time.Now()
	c.Stop()
	log.Printf("notified '%s' at %s paid: %t\n", data.Email, e.NotifiedAt.Format(time.ANSIC), data.Paid)
}
