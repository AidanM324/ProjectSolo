package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/AidanM324/job-queue/internal/job"
	"github.com/AidanM324/job-queue/internal/queue"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: no .env file found")
	}
	conn, ch, err := queue.Connect()
	if err != nil {
		fmt.Println("Error connecting to RabbitMQ:", err)
		return
	}
	defer conn.Close()
	defer ch.Close()

	msgs, err := queue.ConsumeJobs(ch)
	if err != nil {
		fmt.Println("Error consuming jobs:", err)
		return
	}

	fmt.Println("Worker service started, waiting for jobs...")

	for msg := range msgs {
		fmt.Println("Received job:", string(msg.Body))

		var emailJob job.EmailJob
		if err := json.Unmarshal(msg.Body, &emailJob); err != nil {
			fmt.Println("Failed to parse job:", err)
			msg.Ack(false) // bad message, don't retry it forever (we'll improve this in Issue #2's reliability work)
			continue
		}

		if err := job.SendEmail(emailJob); err != nil {
			fmt.Println("Failed to send email:", err)
			continue // leave unacked — RabbitMQ will redeliver it
		}

		fmt.Println("Email sent to:", emailJob.To)

		if err := msg.Ack(false); err != nil {
			fmt.Println("Error acknowledging job:", err)
			continue
		}
		fmt.Println("Job acknowledged")
	}
}