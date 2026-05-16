package services

import (
	"fmt"
	"os"

	"github.com/resendlabs/resend-go"
)

func SendRSVPConfirmation(toEmail, guestName, qrToken string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not set")
	}

	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}

	qrURL := fmt.Sprintf("%s/api/qr/%s.png", backendURL, qrToken)

	client := resend.NewClient(apiKey)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#F7F2EA;font-family:'Helvetica Neue',Arial,sans-serif;">
  <div style="max-width:560px;margin:0 auto;background-color:#ffffff;padding:48px 40px;">

    <div style="text-align:center;margin-bottom:32px;">
      <p style="font-size:11px;letter-spacing:0.3em;text-transform:uppercase;color:#C6A664;margin:0 0 12px;">
        You are invited
      </p>
      <h1 style="font-size:32px;color:#1F1F1F;margin:0;font-weight:400;">
        Faith &amp; Joe
      </h1>
      <div style="width:48px;height:1px;background:#C6A664;margin:20px auto;"></div>
    </div>

    <p style="font-size:16px;color:#1F1F1F;margin:0 0 8px;">
      Dear <strong>%s</strong>,
    </p>
    <p style="font-size:15px;color:rgba(31,31,31,0.65);line-height:1.8;margin:0 0 32px;">
      Your RSVP has been received! We are so excited to celebrate this special
      day with you. Your personal entry QR code is attached to this email.
      Save it and show it at the entrance on the day.
    </p>

    <div style="background:#F7F2EA;padding:24px;text-align:center;margin-bottom:32px;">
      <p style="font-size:13px;color:rgba(31,31,31,0.6);margin:0 0 8px;">
        Your QR code is attached to this email
      </p>
      <p style="font-size:11px;letter-spacing:0.2em;text-transform:uppercase;color:#C6A664;margin:0;">
        entry-qr-code.png
      </p>
    </div>

    <div style="border-top:1px solid #EDE4D3;padding-top:24px;text-align:center;">
      <p style="font-size:13px;color:rgba(31,31,31,0.5);line-height:1.8;margin:0;">
        Saturday, 15 November 2025<br/>
        Ceremony: 2:00 PM | Reception: 5:00 PM<br/>
        Victoria Island, Lagos
      </p>
      <p style="font-size:11px;letter-spacing:0.2em;text-transform:uppercase;color:#C6A664;margin:16px 0 0;">
        #FaithAndJoe2026
      </p>
    </div>

  </div>
</body>
</html>
`, guestName)

	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{toEmail},
		Subject: "Your RSVP is confirmed! | Faith & Joe's Wedding",
		Html:    htmlBody,
		Attachments: []resend.Attachment{
			{
				Filename: "entry-qr-code.png",
				Path:     qrURL,
			},
		},
	}

	_, err := client.Emails.Send(params)
	return err
}
