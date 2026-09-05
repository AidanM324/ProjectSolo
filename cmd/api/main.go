package main

import (
	"fmt"
	"io"
	"net/http"

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

	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		if err := queue.PublishJob(ch, body); err != nil {
			http.Error(w, "failed to publish job", http.StatusInternalServerError)
			return
		}
		fmt.Println("Published job:", string(body))
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w, "Job published")
	})

	fmt.Println("API service listening on :8080")
	http.ListenAndServe(":8080", nil)
}