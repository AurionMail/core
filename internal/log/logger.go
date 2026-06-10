package log

import (
    "log/slog"
    "os"
)

func New(env string) *slog.Logger {
    if env == "dev" {
        return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        }))
    }

    return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
}
