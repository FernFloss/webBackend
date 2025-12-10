package rabbit

import (
	"context"
	"errors"
	"fmt"
	"log"
	"web_backend_v2/config"
	"web_backend_v2/forms"
	"web_backend_v2/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	rabbitConn *amqp.Connection
	rabbitChan *amqp.Channel
)

// InitRabbitMQ establishes connection and channel.
func InitRabbitMQ(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("rabbit config is nil")
	}

	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Prefetch 1 message per consumer to avoid flooding.
	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	rabbitConn = conn
	rabbitChan = ch

	log.Println("Connected to RabbitMQ")
	return nil
}

// CloseRabbitMQ closes channel and connection.
func CloseRabbitMQ() error {
	var errs []error
	if rabbitChan != nil {
		if err := rabbitChan.Close(); err != nil {
			errs = append(errs, fmt.Errorf("channel close: %w", err))
		}
	}
	if rabbitConn != nil {
		if err := rabbitConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("connection close: %w", err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// StartConsumer consumes messages and forwards them to handler until context is canceled.
func StartConsumer(ctx context.Context, queueName string, handler func([]byte) error) error {
	if rabbitChan == nil {
		return fmt.Errorf("RabbitMQ channel is not initialized")
	}

	queue, err := rabbitChan.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	deliveries, err := rabbitChan.Consume(
		queue.Name,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("deliveries channel closed")
			}

			if err := handler(msg.Body); err != nil {
				log.Printf("handler error, delivery will be %s: %v", ackAction(err), err)
				if isNonRetryable(err) {
					msg.Nack(false, false)
				} else {
					msg.Nack(false, true)
				}
				continue
			}

			if err := msg.Ack(false); err != nil {
				log.Printf("failed to ack message: %v", err)
			}
		}
	}
}

func isNonRetryable(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, forms.ErrInvalidCameraEvent) ||
		errors.Is(err, models.ErrCameraNotFound) ||
		errors.Is(err, models.ErrCameraNotAttached)
}

func ackAction(err error) string {
	if isNonRetryable(err) {
		return "acked without requeue"
	}
	return "requeued"
}

