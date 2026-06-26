package main

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fws "github.com/gofiber/websocket/v2"
	"github.com/shadovxw/monopoly/internal/admin"
	"github.com/shadovxw/monopoly/internal/auth"
	"github.com/shadovxw/monopoly/internal/config"
	"github.com/shadovxw/monopoly/internal/db"
	"github.com/shadovxw/monopoly/internal/room"
	"github.com/shadovxw/monopoly/internal/store"
	"github.com/shadovxw/monopoly/internal/ws"
)

func main() {
	cfg := config.Load()

	if err := db.Connect(cfg.DatabaseURL); err != nil {
		log.Fatal("connect db:", err)
	}

	_, filename, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	migrationsDir := filepath.Join(repoRoot, "migrations")

	if err := db.Migrate(migrationsDir); err != nil {
		log.Fatal("migrate:", err)
	}

	rules, err := store.LoadRuleSet(db.DB)
	if err != nil {
		log.Fatal("load ruleset:", err)
	}

	manager := room.NewManager(rules, db.DB)
	validator := auth.NewValidator(cfg.BastionURL)
	wsHandler := ws.NewHandler(manager, validator, db.DB)
	adminHandler := admin.NewHandler(db.DB)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowCredentials: true,
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// Room creation REST endpoint
	app.Post("/rooms", func(c *fiber.Ctx) error {
		r, err := manager.Create()
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"roomCode": r.ID})
	})

	// WebSocket
	app.Use("/ws/:roomCode", wsHandler.Upgrade)
	app.Get("/ws/:roomCode", fws.New(wsHandler.Connect))

	// Admin
	adminHandler.RegisterRoutes(app, cfg.AdminSecret)

	log.Printf("monopoly server listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
