package services

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"wedding-backend/repository"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

type RSVPInput struct {
	FullName   string  `json:"full_name"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	Attendance string  `json:"attendance"`
	GuestCount int     `json:"guest_count"`
	Dietary    *string `json:"dietary"`
}

type RSVPResponse struct {
	RSVP    *repository.RSVP `json:"rsvp"`
	QRImage []byte           `json:"qr_image"`
}

func SubmitRSVP(input RSVPInput) (*RSVPResponse, error) {
	// Check for duplicate
	existing, err := repository.GetRSVPByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("an RSVP already exists for this email address")
	}

	// Generate unique QR token
	qrToken := uuid.New().String()

	rsvp := &repository.RSVP{
		FullName:   input.FullName,
		Email:      input.Email,
		Phone:      input.Phone,
		Attendance: input.Attendance,
		GuestCount: input.GuestCount,
		Dietary:    input.Dietary,
		QRToken:    &qrToken,
	}

	// Save to DB
	if err := repository.CreateRSVP(rsvp); err != nil {
		return nil, fmt.Errorf("failed to save RSVP: %v", err)
	}

	// Generate QR code image
	qrData := fmt.Sprintf("WEDDING-VERIFY:%s", qrToken)
	qrImage, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %v", err)
	}

	// Save QR image to disk
	if err := saveQRImage(qrToken, qrImage); err != nil {
		fmt.Printf("Warning: failed to save QR image: %v\n", err)
	}

	// Send confirmation email (non-blocking)
	// Send confirmation email (non-blocking)
	qrBase64 := base64.StdEncoding.EncodeToString(qrImage)
	go func() {
		if err := SendRSVPConfirmation(input.Email, input.FullName, qrBase64); err != nil {
			fmt.Printf("Failed to send email to %s: %v\n", input.Email, err)
		} else {
			fmt.Printf("Confirmation email sent to %s\n", input.Email)
		}
	}()

	return &RSVPResponse{
		RSVP:    rsvp,
		QRImage: qrImage,
	}, nil
}

func saveQRImage(token string, data []byte) error {
	dir := "qr_images"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, token+".png")
	return os.WriteFile(path, data, 0644)
}

func VerifyQR(token string) (*repository.RSVP, error) {
	rsvp, err := repository.GetRSVPByQRToken(token)
	if err != nil {
		return nil, err
	}
	if rsvp == nil {
		return nil, fmt.Errorf("invalid QR code")
	}
	if rsvp.QRVerified {
		return nil, fmt.Errorf("this QR code has already been used")
	}

	if err := repository.VerifyQRToken(token); err != nil {
		return nil, err
	}

	rsvp.QRVerified = true
	return rsvp, nil
}
