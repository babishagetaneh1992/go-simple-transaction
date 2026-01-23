package email

import (
	"context"
	"log"
)

type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) Notify(ctx context.Context, msg string) error {
	log.Println("📧 EMAIL:", msg)
	return nil
}
