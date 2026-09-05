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