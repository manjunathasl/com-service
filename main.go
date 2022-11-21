package main

import (
	"com-service/internal/config"
	"com-service/internal/scheduler"
	"com-service/internal/service"
)

func main() {
	cfg := config.New()

	customers := service.ReadCSVFile(cfg.CustomerFilePath)
	customerSchedules := service.ComposeSchedules(customers)

	cronJob := scheduler.New(customerSchedules, cfg.ComPubService)

	cronJob.Start()
}
