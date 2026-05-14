package repository

import (
	"database/sql"
	"wedding-backend/config"
)

type RSVP struct {
	ID         string  `json:"id"`
	FullName   string  `json:"full_name"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	Attendance string  `json:"attendance"`
	GuestCount int     `json:"guest_count"`
	Dietary    *string `json:"dietary"`
	QRToken    *string `json:"qr_token"`
	QRVerified bool    `json:"qr_verified"`
	CreatedAt  string  `json:"created_at"`
}

func CreateRSVP(r *RSVP) error {
	query := `
		INSERT INTO rsvps (full_name, email, phone, attendance, guest_count, dietary, qr_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return config.DB.QueryRow(
		query,
		r.FullName,
		r.Email,
		r.Phone,
		r.Attendance,
		r.GuestCount,
		r.Dietary,
		r.QRToken,
	).Scan(&r.ID, &r.CreatedAt)
}

func GetRSVPByEmail(email string) (*RSVP, error) {
	query := `SELECT id, full_name, email, phone, attendance, guest_count, dietary, qr_token, qr_verified, created_at FROM rsvps WHERE email = $1`
	r := &RSVP{}
	err := config.DB.QueryRow(query, email).Scan(
		&r.ID, &r.FullName, &r.Email, &r.Phone,
		&r.Attendance, &r.GuestCount, &r.Dietary,
		&r.QRToken, &r.QRVerified, &r.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func GetRSVPByQRToken(token string) (*RSVP, error) {
	query := `SELECT id, full_name, email, attendance, guest_count, qr_verified FROM rsvps WHERE qr_token = $1`
	r := &RSVP{}
	err := config.DB.QueryRow(query, token).Scan(
		&r.ID, &r.FullName, &r.Email,
		&r.Attendance, &r.GuestCount, &r.QRVerified,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func VerifyQRToken(token string) error {
	query := `UPDATE rsvps SET qr_verified = TRUE WHERE qr_token = $1`
	_, err := config.DB.Exec(query, token)
	return err
}
