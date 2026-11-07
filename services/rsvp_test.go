package services

import (
	"bytes"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"wedding-backend/config"

	"github.com/DATA-DOG/go-sqlmock"
)

var rsvpCols = []string{"id", "full_name", "email", "phone", "attendance", "guest_count", "dietary", "qr_token", "qr_verified", "created_at"}
var qrCols = []string{"id", "full_name", "email", "attendance", "guest_count", "qr_verified"}

func useMockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	config.DB = db
	t.Cleanup(func() { db.Close() })
	return mock
}

// SubmitRSVP writes the QR image to ./qr_images, so run it in a temp dir
// to keep the repository clean.
func inTempDir(t *testing.T) {
	t.Helper()
	old, _ := os.Getwd()
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

func validInput() RSVPInput {
	return RSVPInput{
		FullName: "Grace Kalu", Email: "grace@example.com", Phone: "08012345678",
		Attendance: "yes", GuestCount: 2,
	}
}

func TestSubmitRSVP_RejectsDuplicateEmail(t *testing.T) {
	mock := useMockDB(t)
	mock.ExpectQuery("SELECT .* FROM rsvps WHERE email").
		WithArgs("grace@example.com").
		WillReturnRows(sqlmock.NewRows(rsvpCols).AddRow("id-1", "Grace Kalu", "grace@example.com", "08012345678", "yes", 2, nil, "tok", false, "2026-01-01"))

	res, err := SubmitRSVP(validInput())

	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("err = %v, want duplicate-email error", err)
	}
	if res != nil {
		t.Error("expected nil result for a duplicate RSVP")
	}
	// The insert must never run for a duplicate (ExpectationsWereMet fails on unexpected calls).
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSubmitRSVP_CreatesRSVPWithTokenAndQRImage(t *testing.T) {
	inTempDir(t)
	mock := useMockDB(t)
	mock.ExpectQuery("SELECT .* FROM rsvps WHERE email").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("INSERT INTO rsvps").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("new-id", "2026-01-01T00:00:00Z"))

	res, err := SubmitRSVP(validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RSVP.ID != "new-id" {
		t.Errorf("ID = %q, want new-id", res.RSVP.ID)
	}
	if res.RSVP.QRToken == nil || len(*res.RSVP.QRToken) != 36 {
		t.Errorf("QR token should be a UUID, got %v", res.RSVP.QRToken)
	}
	pngHeader := []byte{0x89, 'P', 'N', 'G'}
	if !bytes.HasPrefix(res.QRImage, pngHeader) {
		t.Error("QR image is not a PNG")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSubmitRSVP_GeneratesUniqueTokens(t *testing.T) {
	inTempDir(t)
	tokens := map[string]bool{}
	for i := 0; i < 3; i++ {
		mock := useMockDB(t)
		mock.ExpectQuery("SELECT .* FROM rsvps WHERE email").WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery("INSERT INTO rsvps").
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("id", "now"))
		res, err := SubmitRSVP(validInput())
		if err != nil {
			t.Fatal(err)
		}
		tokens[*res.RSVP.QRToken] = true
	}
	if len(tokens) != 3 {
		t.Errorf("got %d unique tokens from 3 RSVPs, want 3", len(tokens))
	}
}

func TestSubmitRSVP_ReturnsErrorWhenInsertFails(t *testing.T) {
	inTempDir(t)
	mock := useMockDB(t)
	mock.ExpectQuery("SELECT .* FROM rsvps WHERE email").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("INSERT INTO rsvps").WillReturnError(errors.New("connection reset"))

	_, err := SubmitRSVP(validInput())
	if err == nil || !strings.Contains(err.Error(), "failed to save RSVP") {
		t.Fatalf("err = %v, want wrapped save error", err)
	}
}

func TestVerifyQR_UnknownTokenIsInvalid(t *testing.T) {
	mock := useMockDB(t)
	mock.ExpectQuery("SELECT .* FROM rsvps WHERE qr_token").WillReturnError(sql.ErrNoRows)

	_, err := VerifyQR("does-not-exist")
	if err == nil || err.Error() != "invalid QR code" {
		t.Fatalf("err = %v, want 'invalid QR code'", err)
	}
}

func TestVerifyQR_RejectsAlreadyUsedToken(t *testing.T) {
	mock := useMockDB(t)
	mock.ExpectQuery("SELECT .* FROM rsvps WHERE qr_token").
		WillReturnRows(sqlmock.NewRows(qrCols).AddRow("id-1", "Grace Kalu", "g@example.com", "yes", 2, true))

	_, err := VerifyQR("used-token")
	if err == nil || !strings.Contains(err.Error(), "already been used") {
		t.Fatalf("err = %v, want already-used error", err)
	}
	// No UPDATE should have been attempted.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestVerifyQR_MarksTokenUsedOnFirstScan(t *testing.T) {
	mock := useMockDB(t)
	mock.ExpectQuery("SELECT .* FROM rsvps WHERE qr_token").
		WithArgs("good-token").
		WillReturnRows(sqlmock.NewRows(qrCols).AddRow("id-1", "Grace Kalu", "g@example.com", "yes", 2, false))
	mock.ExpectExec("UPDATE rsvps SET qr_verified = TRUE").
		WithArgs("good-token").
		WillReturnResult(sqlmock.NewResult(0, 1))

	guest, err := VerifyQR("good-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !guest.QRVerified {
		t.Error("returned guest should be marked verified")
	}
	if guest.FullName != "Grace Kalu" {
		t.Errorf("FullName = %q", guest.FullName)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestVerifyQR_PropagatesDatabaseErrors(t *testing.T) {
	mock := useMockDB(t)
	mock.ExpectQuery("SELECT .* FROM rsvps WHERE qr_token").WillReturnError(errors.New("db down"))

	_, err := VerifyQR("any")
	if err == nil || err.Error() != "db down" {
		t.Fatalf("err = %v, want db down", err)
	}
}
