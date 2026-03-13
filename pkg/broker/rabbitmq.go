package broker

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("pkg: failed to connect to RabbitMQ server: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("pkg: failed to open RabbitMQ channel: %w", err)
	}

	return &RabbitMQClient{
		connection: connection,
		channel:    channel,
	}, nil
}

func (r *RabbitMQClient) DeclareExchange(name, kind string) error {
	err := r.channel.ExchangeDeclare(
		name,
		kind,
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		return fmt.Errorf("pkg: failed to declare RabbitMQ exchange: %w", err)
	}
	return nil
}

func (r *RabbitMQClient) DeclareQueue(name string) (*amqp.Queue, error) {
	queue, err := r.channel.QueueDeclare(name, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("pkg: failed to declare RabbitMQ queue: %w", err)
	}

	return &queue, nil
}

func (r *RabbitMQClient) BindQueue(queueName, exchange, routingKey string) error {
	err := r.channel.QueueBind(queueName, routingKey, exchange, false, nil)
	if err != nil {
		return fmt.Errorf("pkg: failed to bind queue to exchange: %w", err)
	}

	return nil
}

func (r *RabbitMQClient) Publish(ctx context.Context, exchange, routingKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("pkg: failed to marshal payload: %w", err)
	}

	err = r.channel.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
	if err != nil {
		return fmt.Errorf("pkg: failed to publish to RabbitMQ exchange: %w", err)
	}

	return nil
}

func (r *RabbitMQClient) Consume(ctx context.Context, queueName string) (<-chan amqp.Delivery, error) {
	result, err := r.channel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("pkg: failed to consume RabbitMQ queue: %w", err)
	}

	return result, nil
}

func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.connection != nil {
		return r.connection.Close()
	}
	return nil
}
