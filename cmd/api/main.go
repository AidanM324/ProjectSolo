package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/joho/godotenv"

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

	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		id, err := store.InsertJob(r.Context(), db, "send_email", body)
		if err != nil {
			fmt.Println("Failed to insert job:", err)
			http.Error(w, "failed to save job", http.StatusInternalServerError)
			return
		}

		wrapped, err := json.Marshal(queuedMessage{ID: id, Payload: body})
		if err != nil {
			http.Error(w, "failed to prepare job", http.StatusInternalServerError)
			return
		}

		if err := queue.PublishJob(ch, wrapped); err != nil {
			http.Error(w, "failed to publish job", http.StatusInternalServerError)
			return
		}

		fmt.Println("Job created with ID:", id)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w, "Job created:", id)
	})

	fmt.Println("API service listening on :8080")
	http.ListenAndServe(":8080", nil)
}