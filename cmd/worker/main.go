package main

import (
	"fmt"

	"github.com/AidanM324/job-queue/internal/queue"
)

func main() {
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

		// TODO (Issue #5): actually execute the job (send an email)

		if err := msg.Ack(false); err != nil {
			fmt.Println("Error acknowledging job:", err)
			continue
		}
		fmt.Println("Job acknowledged")
	}
}