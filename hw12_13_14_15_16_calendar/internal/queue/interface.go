package queue

import "context"

type Producer interface {
	SendMessage(ctx context.Context, topic string, key string, message []byte) error
	Close() error
}

type Consumer interface {
	ConsumeMessages(ctx context.Context, topic string, handler MessageHandler) error
	Close() error
}

type MessageHandler func(ctx context.Context, message []byte) error
