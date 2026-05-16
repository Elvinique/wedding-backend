# Wedding App — Go Backend

A production-ready REST API built with Go and Fiber, powering a luxury Nigerian wedding invitation web app. Handles RSVP submissions, QR code generation, guestbook messages, and email confirmations.

---

## Tech Stack

- **Language:** Go 1.22+
- **Framework:** Fiber v2
- **Database:** PostgreSQL via Supabase
- **Email:** Resend
- **QR Code:** go-qrcode
- **Hosting:** Render

---

## Features

- RSVP submission with duplicate prevention
- Unique QR code generation per guest
- QR code verification endpoint
- Email confirmation with QR attachment
- Guestbook message submission and retrieval
- Rate limiting (60 requests/min)
- CORS protection
- Environment-based configuration

---

## Project Structure

```
wedding-backend/
├── config/
│   └── database.go        # Database connection
├── handlers/
│   ├── rsvp.go            # RSVP HTTP handlers
│   └── guestbook.go       # Guestbook HTTP handlers
├── repository/
│   ├── rsvp.go            # RSVP database queries
│   └── guestbook.go       # Guestbook database queries
├── services/
│   ├── rsvp.go            # RSVP business logic + QR generation
│   └── email.go           # Email confirmation via Resend
├── qr_images/             # Generated QR code PNG files
├── .env                   # Environment variables (not committed)
├── .gitignore
├── go.mod
├── go.sum
└── main.go                # Entry point, routes, middleware
```

---

## API Endpoints

### Health
```
GET /health
```
Returns service status.

---

### RSVP
```
POST /api/rsvp
```
Submit a guest RSVP. Generates a unique QR code and sends a confirmation email.

**Request body:**
```json
{
  "full_name": "Grace Kalu",
  "email": "grace@example.com",
  "phone": "08012345678",
  "attendance": "yes",
  "guest_count": 2,
  "dietary": "vegetarian"
}
```

**Response:**
```json
{
  "message": "RSVP submitted successfully",
  "rsvp": { ... },
  "qr_code": "<base64 PNG>"
}
```

---

```
GET /api/rsvp/verify/:token
```
Verify a guest QR token at the entrance. Marks it as used.

---

### Guestbook
```
GET /api/guestbook?limit=20&offset=0
```
Fetch paginated approved guestbook messages.

---

```
POST /api/guestbook
```
Submit a congratulatory message.

**Request body:**
```json
{
  "name": "Aunty Ngozi",
  "message": "Wishing you both a lifetime of love!"
}
```

---

### QR Images
```
GET /api/qr/:token.png
```
Serves the generated QR code image for a guest.

---

## Database Schema

```sql
CREATE TABLE rsvps (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  full_name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  phone TEXT NOT NULL,
  attendance TEXT NOT NULL CHECK (attendance IN ('yes', 'no')),
  guest_count INTEGER NOT NULL DEFAULT 1,
  dietary TEXT,
  qr_token TEXT UNIQUE,
  qr_verified BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE guestbook (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  name TEXT NOT NULL,
  message TEXT NOT NULL,
  approved BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## Environment Variables

Create a `.env` file in the project root:

```env
PORT=8080
DATABASE_URL=postgresql://postgres.[ref]:[password]@[host]:6543/postgres
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-role-key
RESEND_API_KEY=re_your_resend_key
FRONTEND_URL=https://your-frontend.vercel.app
BACKEND_URL=https://your-backend.onrender.com
```

> ⚠️ Never commit `.env` to version control.

---

## Running Locally

**Prerequisites:** Go 1.22+, PostgreSQL (or Supabase account)

```bash
# Clone the repository
git clone https://github.com/Elvinique/wedding-backend.git
cd wedding-backend

# Install dependencies
go mod tidy

# Set up environment variables
cp .env.example .env
# Edit .env with your credentials

# Run the server
go run main.go
```

Server starts at `http://localhost:8080`.

---

## Building for Production

```bash
go build -o main .
./main
```

---

## Deployment

This backend is deployed on **Render** as a Web Service.

| Setting | Value |
|---|---|
| Runtime | Go |
| Build Command | `go build -o main .` |
| Start Command | `./main` |
| Region | Frankfurt (EU Central) |

Set all environment variables in Render's dashboard under **Environment**.

---

## Related

- **Frontend:** [wedding-app](https://github.com/Elvinique/wedding-app) — Next.js on Vercel
- **Database:** Supabase (PostgreSQL)
- **Live API:** https://wedding-backend-zsdn.onrender.com

---

## License

MIT
