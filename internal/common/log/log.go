package log

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func init() {
	logFile, err := os.Create("log.txt")
	if err != nil {
		panic(err)
	}

	multi := zerolog.MultiLevelWriter(logFile, os.Stdout)

	logger = zerolog.New(multi).With().Timestamp().Logger()
}

func LoggerFromContext(ctx context.Context) *zerolog.Logger {
	ctx = logger.WithContext(ctx)

	return zerolog.Ctx(ctx)
}
