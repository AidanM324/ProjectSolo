package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joho/godotenv"

	"github.com/AidanM324/job-queue/internal/job"
	"github.com/AidanM324/job-queue/internal/queue"
	"github.com/AidanM324/job-queue/internal/store"
)

type queuedMessage struct {
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: no .env file found")
	}

	ctx := context.Background()

	conn, ch, err := queue.Connect()
	if err != nil {
		fmt.Println("Error connecting to RabbitMQ:", err)
		return
	}
	defer conn.Close()
	defer ch.Close()

	db, err := store.Connect(ctx)
	if err != nil {
		fmt.Println("Error connecting to Postgres:", err)
		return
	}
	defer db.Close()

	msgs, err := queue.ConsumeJobs(ch)
	if err != nil {
		fmt.Println("Error consuming jobs:", err)
		return
	}

	fmt.Println("Worker service started, waiting for jobs...")

	for msg := range msgs {
		var qm queuedMessage
		if err := json.Unmarshal(msg.Body, &qm); err != nil {
			fmt.Println("Failed to parse message:", err)
			msg.Ack(false)
			continue
		}

		fmt.Println("Processing job:", qm.ID)
		store.UpdateJobStatus(ctx, db, qm.ID, "running")

		var emailJob job.EmailJob
		if err := json.Unmarshal(qm.Payload, &emailJob); err != nil {
			fmt.Println("Failed to parse email job:", err)
			store.UpdateJobStatus(ctx, db, qm.ID, "failed")
			msg.Ack(false)
			continue
		}

		if err := job.SendEmail(emailJob); err != nil {
			fmt.Println("Failed to send email:", err)
			store.UpdateJobStatus(ctx, db, qm.ID, "failed")
			continue
		}

		fmt.Println("Email sent to:", emailJob.To)
		store.UpdateJobStatus(ctx, db, qm.ID, "done")

		if err := msg.Ack(false); err != nil {
			fmt.Println("Error acknowledging job:", err)
			continue
		}
		fmt.Println("Job acknowledged")
	}
}