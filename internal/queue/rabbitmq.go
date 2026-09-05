package queue

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/") // Connect to RabbitMQ server
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	ch, err := conn.Channel() // Open a channel on AMQP connection
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open channel: %w", err)
	}
	_, err = ch.QueueDeclare( // Declare a queue if it doesn't exist
		"jobs", // queue name
		true,   // durable (survives RabbitMQ restart)  marking it "durable" so messages survive a RabbitMQ restart (important for a reliability-focused project like this).
		false,  // auto-delete
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to declare queue: %w", err)
	}
	return conn, ch, nil
}

func PublishJob(ch *amqp.Channel, body []byte) error {
	err := ch.Publish(
		"",     // exchange (default)
		"jobs", // routing key = queue name
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // survive broker restart
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish job: %w", err)
	}
	return nil
}

func ConsumeJobs(ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume(
		"jobs", // queue name
		"",     // consumer tag (auto-generated)
		false,  // auto-ack (false = manual acknowledgement) RabbitMQ will NOT remove a message from the queue just because a worker received it — it stays in the queue until the worker explicitly acknowledges ("acks") it
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}
	return msgs, nil
}

//The key detail here is `auto-ack: false`. 
//This means RabbitMQ will NOT remove a message from the queue just because a worker received it — it stays in the queue until the worker explicitly acknowledges ("acks") it. This is deliberate and important: if the worker crashes mid-job, 
//the unacknowledged message goes back to the queue for another worker to pick up, instead of being silently lost. This is the foundation of the reliability work we'll build on in Issue #2 (retries/dead-letter queue).

//`ch.Consume` returns a Go channel (`<-chan amqp.Delivery`) — not to be confused with a RabbitMQ channel, this is Go's built-in concurrency primitive for passing values between goroutines. 
// Each `amqp.Delivery` that comes through represents one message pulled from the queue, and it has a `.Body` field (the raw bytes you published) and an `.Ack()` method you call once you've successfully processed it.