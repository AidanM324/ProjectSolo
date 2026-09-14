package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joho/godotenv"
	amqp091 "github.com/rabbitmq/amqp091-go"

	"github.com/AidanM324/job-queue/internal/job"
	"github.com/AidanM324/job-queue/internal/queue"
	"github.com/AidanM324/job-queue/internal/store"
)

type queuedMessage struct {
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

const maxRetries = 5

func getRetryCount(msg amqp091.Delivery) int {
	if val, ok := msg.Headers["x-retry-count"]; ok {
		if count, ok := val.(int32); ok {
			return int(count)
		}
	}
	return 0
}

func backoffDelay(retryCount int) int {
	// 2s, 4s, 8s, 16s, 32s ... in milliseconds
	return 2000 * (1 << retryCount)
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

	if err := queue.DeclareRetryQueues(ch); err != nil {
		fmt.Println("Error declaring retry queues:", err)
		return
	}

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
			retryCount := getRetryCount(msg)
			fmt.Printf("Failed to send email (attempt %d): %v\n", retryCount+1, err)

			if retryCount >= maxRetries {
				fmt.Println("Max retries exceeded, sending to dead-letter queue:", qm.ID)
				store.UpdateJobStatus(ctx, db, qm.ID, "dead")
				queue.PublishToDeadLetter(ch, msg.Body)
				msg.Ack(false)
				continue
			}

			delay := backoffDelay(retryCount)
			fmt.Printf("Retrying in %dms\n", delay)
			store.UpdateJobStatus(ctx, db, qm.ID, "failed")
			queue.RepublishWithDelay(ch, msg.Body, retryCount+1, delay)
			msg.Ack(false) // remove from main queue, it now lives in jobs_retry
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
