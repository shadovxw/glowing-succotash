package admin

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes mounts all admin endpoints under /admin.
func (h *Handler) RegisterRoutes(app *fiber.App, adminSecret string) {
	g := app.Group("/admin", requireAdminKey(adminSecret))

	g.Get("/config", h.GetConfig)
	g.Put("/config", h.UpdateConfig)

	g.Get("/tiles", h.GetTiles)
	g.Put("/tiles/:position", h.UpdateTile)

	g.Get("/properties", h.GetProperties)
	g.Put("/properties/:id", h.UpdateProperty)

	g.Get("/cards", h.GetCards)
	g.Post("/cards", h.CreateCard)
	g.Put("/cards/:id", h.UpdateCard)
	g.Delete("/cards/:id", h.DeleteCard)

	g.Get("/players", h.GetPlayers)
	g.Get("/players/:id/stats", h.GetPlayerStats)
}

// ---- Config ----

func (h *Handler) GetConfig(c *fiber.Ctx) error {
	var row map[string]any
	if err := h.db.Table("game_config").First(&row).Error; err != nil {
		return err
	}
	return c.JSON(row)
}

func (h *Handler) UpdateConfig(c *fiber.Ctx) error {
	var body map[string]any
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}
	delete(body, "id") // prevent overwriting the singleton key
	if err := h.db.Table("game_config").Where("id = 1").Updates(body).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Tiles ----

func (h *Handler) GetTiles(c *fiber.Ctx) error {
	var rows []map[string]any
	if err := h.db.Table("tiles").Order("position").Find(&rows).Error; err != nil {
		return err
	}
	return c.JSON(rows)
}

func (h *Handler) UpdateTile(c *fiber.Ctx) error {
	pos := c.Params("position")
	var body map[string]any
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}
	// Protect fixed fields
	delete(body, "position")
	delete(body, "type")
	if err := h.db.Table("tiles").Where("position = ?", pos).Updates(body).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Properties ----

func (h *Handler) GetProperties(c *fiber.Ctx) error {
	var rows []map[string]any
	if err := h.db.Table("properties").
		Joins("JOIN tiles ON tiles.position = properties.tile_position").
		Select("properties.*, tiles.name, tiles.color_group").
		Find(&rows).Error; err != nil {
		return err
	}
	return c.JSON(rows)
}

func (h *Handler) UpdateProperty(c *fiber.Ctx) error {
	id := c.Params("id")
	var body map[string]any
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}
	delete(body, "id")
	delete(body, "tile_position")
	if err := h.db.Table("properties").Where("id = ?", id).Updates(body).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Cards ----

func (h *Handler) GetCards(c *fiber.Ctx) error {
	query := h.db.Table("cards").Order("deck, sort_order")
	if deck := c.Query("deck"); deck != "" {
		query = query.Where("deck = ?", deck)
	}
	var rows []map[string]any
	if err := query.Find(&rows).Error; err != nil {
		return err
	}
	return c.JSON(rows)
}

func (h *Handler) CreateCard(c *fiber.Ctx) error {
	var body map[string]any
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.db.Table("cards").Create(body).Error; err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(body)
}

func (h *Handler) UpdateCard(c *fiber.Ctx) error {
	id := c.Params("id")
	var body map[string]any
	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}
	delete(body, "id")
	if err := h.db.Table("cards").Where("id = ?", id).Updates(body).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteCard(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.db.Table("cards").Where("id = ?", id).Delete(nil).Error; err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Players ----

func (h *Handler) GetPlayers(c *fiber.Ctx) error {
	var rows []map[string]any
	if err := h.db.Table("players").Order("games_played DESC").Find(&rows).Error; err != nil {
		return err
	}
	return c.JSON(rows)
}

func (h *Handler) GetPlayerStats(c *fiber.Ctx) error {
	id := c.Params("id")
	var row map[string]any
	if err := h.db.Table("players").Where("bastion_user_id = ?", id).First(&row).Error; err != nil {
		return fiber.ErrNotFound
	}
	return c.JSON(row)
}

// ---- Middleware ----

func requireAdminKey(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if secret != "" && c.Get("X-Admin-Secret") != secret {
			return fiber.ErrUnauthorized
		}
		return c.Next()
	}
}
