package handlers

import (
	"encoding/base64"
	"wedding-backend/config"
	"wedding-backend/repository"
	"wedding-backend/services"

	"github.com/gofiber/fiber/v2"
)

func SubmitRSVP(c *fiber.Ctx) error {
	var input services.RSVPInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Basic validation
	if input.FullName == "" || len(input.FullName) < 2 {
		return c.Status(400).JSON(fiber.Map{"error": "Full name is required"})
	}
	if input.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email is required"})
	}
	if input.Phone == "" || len(input.Phone) < 10 {
		return c.Status(400).JSON(fiber.Map{"error": "Valid phone number is required"})
	}
	if input.Attendance != "yes" && input.Attendance != "no" {
		return c.Status(400).JSON(fiber.Map{"error": "Attendance must be yes or no"})
	}
	if input.GuestCount < 1 || input.GuestCount > 5 {
		return c.Status(400).JSON(fiber.Map{"error": "Guest count must be between 1 and 5"})
	}

	result, err := services.SubmitRSVP(input)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "RSVP submitted successfully",
		"rsvp":    result.RSVP,
		"qr_code": base64.StdEncoding.EncodeToString(result.QRImage),
	})
}

func VerifyQR(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Token is required"})
	}

	rsvp, err := services.VerifyQR(token)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "QR code verified successfully",
		"guest":   rsvp,
	})
}
func GetAllRSVPs(c *fiber.Ctx) error {
	query := `
		SELECT id, full_name, email, phone, attendance, guest_count, dietary, qr_token, qr_verified, created_at
		FROM rsvps
		ORDER BY created_at DESC
	`
	rows, err := config.DB.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch RSVPs"})
	}
	defer rows.Close()

	var rsvps []repository.RSVP
	for rows.Next() {
		var r repository.RSVP
		if err := rows.Scan(
			&r.ID, &r.FullName, &r.Email, &r.Phone,
			&r.Attendance, &r.GuestCount, &r.Dietary,
			&r.QRToken, &r.QRVerified, &r.CreatedAt,
		); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to scan RSVPs"})
		}
		rsvps = append(rsvps, r)
	}

	if rsvps == nil {
		rsvps = []repository.RSVP{}
	}

	return c.JSON(fiber.Map{"rsvps": rsvps})
}
