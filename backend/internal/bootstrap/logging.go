package bootstrap

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/luizf/go-chat/backend/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogging(cfg config.Config) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	level, err := zerolog.ParseLevel(strings.ToLower(cfg.LogLevel))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var out io.Writer = os.Stdout
	if cfg.LogFormat == "json" {
		out = os.Stdout
	} else {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}
	log.Logger = log.Output(out)
}
