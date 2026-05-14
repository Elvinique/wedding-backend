package handlers

import (
	"wedding-backend/repository"

	"github.com/gofiber/fiber/v2"
)

func GetMessages(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	messages, err := repository.GetMessages(limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch messages"})
	}

	if messages == nil {
		messages = []repository.GuestMessage{}
	}

	return c.JSON(fiber.Map{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
	})
}

func PostMessage(c *fiber.Ctx) error {
	var msg repository.GuestMessage

	if err := c.BodyParser(&msg); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(msg.Name) < 2 {
		return c.Status(400).JSON(fiber.Map{"error": "Name is required"})
	}
	if len(msg.Message) < 5 {
		return c.Status(400).JSON(fiber.Map{"error": "Message is too short"})
	}
	if len(msg.Message) > 200 {
		return c.Status(400).JSON(fiber.Map{"error": "Message is too long"})
	}

	if err := repository.CreateMessage(&msg); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save message"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Message posted successfully",
		"data":    msg,
	})
}
