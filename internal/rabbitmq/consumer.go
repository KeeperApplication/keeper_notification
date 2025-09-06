package rabbitmq

import (
	"context"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"keeper.notifications.go/internal/config"
	"keeper.notifications.go/internal/notification"
)

const (
	exchangeName       = "keeper.exchange"
	notificationsQueue = "keeper.notifications"
	routingKey         = "event.notification.#"
)

type Consumer struct {
	logger              *slog.Logger
	cfg                 *config.Config
	notificationService *notification.Service
}

func NewConsumer(logger *slog.Logger, cfg *config.Config, notificationService *notification.Service) *Consumer {
	return &Consumer{
		logger:              logger,
		cfg:                 cfg,
		notificationService: notificationService,
	}
}

func (s *Consumer) Start(ctx context.Context) {
	s.logger.Info("starting rabbitmq consumer")
	go s.listenForMessages(ctx)
}

func (s *Consumer) listenForMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("context cancelled, stopping consumer")
			return
		default:
			conn, err := amqp.Dial(s.cfg.RabbitMQURL)
			if err != nil {
				s.logger.Error("failed to connect to rabbitmq, retrying in 5s", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}
			defer conn.Close()

			ch, err := conn.Channel()
			if err != nil {
				s.logger.Error("failed to open a channel, retrying", "error", err)
				continue
			}
			defer ch.Close()

			err = ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)
			if err != nil {
				s.logger.Error("failed to declare exchange", "error", err)
				continue
			}

			q, err := ch.QueueDeclare(notificationsQueue, true, false, false, false, nil)
			if err != nil {
				s.logger.Error("failed to declare queue", "error", err)
				continue
			}

			err = ch.QueueBind(q.Name, routingKey, exchangeName, false, nil)
			if err != nil {
				s.logger.Error("failed to bind queue", "error", err)
				continue
			}

			msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
			if err != nil {
				s.logger.Error("failed to register consumer", "error", err)
				continue
			}

			s.logger.Info("rabbitmq consumer connected and waiting for messages")

			for d := range msgs {
				s.notificationService.ProcessNotification(d.Body)
			}
		}
	}
}
