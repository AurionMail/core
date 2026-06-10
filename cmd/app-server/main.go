package main

import (
    "aurion/core/internal/config"
    "aurion/core/internal/http"
    "aurion/core/internal/log"
)

func main() {
    cfg := config.Load()
    logger := log.New()

    router := http.NewRouter(logger)

    logger.Info("Starting Aurion core server", "port", cfg.AppPort)
    router.Run(":" + cfg.AppPort)
}
