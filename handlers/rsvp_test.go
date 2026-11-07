package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"wedding-backend/config"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
)

func newTestApp() *fiber.App {
	app := fiber.New()
	app.Post("/api/rsvp", SubmitRSVP)
	app.Get("/api/admin/rsvps", GetAllRSVPs)
	app.Post("/api/guestbook", PostMessage)
	return app
}

func post(t *testing.T, app *fiber.App, path, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest("POST", path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	return resp.StatusCode, out
}

// Every case below must be rejected by input validation before any
// database access happens, so no database is needed.
func TestSubmitRSVP_ValidationRejectsBadInput(t *testing.T) {
	app := newTestApp()

	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{"malformed JSON", `{not json`, "Invalid request body"},
		{"name too short", `{"full_name":"A","email":"a@b.com","phone":"08012345678","attendance":"yes","guest_count":1}`, "Full name is required"},
		{"missing name", `{"email":"a@b.com","phone":"08012345678","attendance":"yes","guest_count":1}`, "Full name is required"},
		{"missing email", `{"full_name":"Grace Kalu","phone":"08012345678","attendance":"yes","guest_count":1}`, "Email is required"},
		{"phone too short", `{"full_name":"Grace Kalu","email":"a@b.com","phone":"0801","attendance":"yes","guest_count":1}`, "Valid phone number is required"},
		{"bad attendance value", `{"full_name":"Grace Kalu","email":"a@b.com","phone":"08012345678","attendance":"maybe","guest_count":1}`, "Attendance must be yes or no"},
		{"guest count zero", `{"full_name":"Grace Kalu","email":"a@b.com","phone":"08012345678","attendance":"yes","guest_count":0}`, "Guest count must be between 1 and 5"},
		{"guest count above limit", `{"full_name":"Grace Kalu","email":"a@b.com","phone":"08012345678","attendance":"yes","guest_count":6}`, "Guest count must be between 1 and 5"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, body := post(t, app, "/api/rsvp", tc.body)
			if status != 400 {
				t.Fatalf("status = %d, want 400", status)
			}
			if got := body["error"]; got != tc.wantError {
				t.Errorf("error = %q, want %q", got, tc.wantError)
			}
		})
	}
}

func TestPostMessage_ValidationRejectsBadInput(t *testing.T) {
	app := newTestApp()

	tests := []struct {
		name      string
		body      string
		wantError string
	}{
		{"malformed JSON", `{oops`, "Invalid request body"},
		{"name too short", `{"name":"A","message":"Congratulations!"}`, "Name is required"},
		{"message too short", `{"name":"Ngozi","message":"Hi"}`, "Message is too short"},
		{"message too long", `{"name":"Ngozi","message":"` + strings.Repeat("x", 201) + `"}`, "Message is too long"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, body := post(t, app, "/api/guestbook", tc.body)
			if status != 400 {
				t.Fatalf("status = %d, want 400", status)
			}
			if got := body["error"]; got != tc.wantError {
				t.Errorf("error = %q, want %q", got, tc.wantError)
			}
		})
	}
}

// An empty table must serialise as [] (not null) so the frontend can map over it.
func TestGetAllRSVPs_EmptyReturnsEmptyArray(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	config.DB = db

	cols := []string{"id", "full_name", "email", "phone", "attendance", "guest_count", "dietary", "qr_token", "qr_verified", "created_at"}
	mock.ExpectQuery("SELECT .* FROM rsvps").WillReturnRows(sqlmock.NewRows(cols))

	req := httptest.NewRequest("GET", "/api/admin/rsvps", nil)
	resp, err := newTestApp().Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(raw), `"rsvps":[]`) {
		t.Errorf("body = %s, want an empty rsvps array", raw)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
