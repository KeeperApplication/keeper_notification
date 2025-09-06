package notification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
	"keeper.notifications.go/internal/config"
)

type NotificationPayload struct {
	RecipientUsername string `json:"recipientUsername"`
	SenderUsername    string `json:"senderUsername"`
	MessageContent    string `json:"messageContent"`
	RoomID            int64  `json:"roomId"`
	FcmToken          string `json:"fcmToken"`
}

type Service struct {
	logger         *slog.Logger
	firebaseApp    *firebase.App
	firebaseClient *messaging.Client
}

func NewService(logger *slog.Logger, cfg *config.Config) (*Service, error) {
	ctx := context.Background()
	opts := option.WithCredentialsFile(cfg.FCMCredentialsFilePath)

	app, err := firebase.NewApp(ctx, nil, opts)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &Service{
		logger:         logger,
		firebaseApp:    app,
		firebaseClient: client,
	}, nil
}

func (s *Service) ProcessNotification(body []byte) {
	var payload NotificationPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		s.logger.Error("failed to decode notification payload", "error", err)
		return
	}

	s.logger.Info("processing notification", "recipient_hash", hashUsername(payload.RecipientUsername))
	s.sendPushNotification(payload)
}

func (s *Service) sendPushNotification(payload NotificationPayload) {
	if payload.FcmToken == "" {
		s.logger.Warn("fcm token is missing from payload, skipping notification", "recipient_hash", hashUsername(payload.RecipientUsername))
		return
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: "new message from " + payload.SenderUsername,
			Body:  payload.MessageContent,
		},
		Token: payload.FcmToken,
	}

	response, err := s.firebaseClient.Send(context.Background(), message)
	if err != nil {
		s.logger.Error("failed to send fcm message", "error", err, "recipient_hash", hashUsername(payload.RecipientUsername))
		return
	}

	s.logger.Info("successfully sent fcm message", "response", response, "recipient_hash", hashUsername(payload.RecipientUsername))
}

func hashUsername(username string) string {
	h := sha256.New()
	h.Write([]byte(username))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}
