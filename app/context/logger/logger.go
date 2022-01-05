package logger

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

type key int

const (
	contextKey key = iota
)

func Get(ctx context.Context) logrus.FieldLogger {
	logger := ctx.Value(contextKey)
	fieldLogger, ok := logger.(logrus.FieldLogger)
	if !ok || fieldLogger == nil {
		return &logrus.Logger{
			Out: os.Stdout,
		}
	}

	return fieldLogger
}

func WithLogger(ctx context.Context, logger logrus.FieldLogger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}
