package repository

import (
	"wedding-backend/config"
)

type GuestMessage struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	Approved  bool   `json:"approved"`
	CreatedAt string `json:"created_at"`
}

func CreateMessage(m *GuestMessage) error {
	query := `
		INSERT INTO guestbook (name, message)
		VALUES ($1, $2)
		RETURNING id, created_at
	`
	return config.DB.QueryRow(query, m.Name, m.Message).Scan(&m.ID, &m.CreatedAt)
}

func GetMessages(limit, offset int) ([]GuestMessage, error) {
	query := `
		SELECT id, name, message, approved, created_at
		FROM guestbook
		WHERE approved = TRUE
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := config.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []GuestMessage
	for rows.Next() {
		var m GuestMessage
		if err := rows.Scan(&m.ID, &m.Name, &m.Message, &m.Approved, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
