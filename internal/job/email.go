package job

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

type EmailJob struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func SendEmail(e EmailJob) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	auth := smtp.PlainAuth("", username, password, host)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", e.To, e.Subject, e.Body))
	addr := fmt.Sprintf("%s:%s", host, port)
	err := smtp.SendMail(addr, auth, "noreply@jobqueue.dev", []string{e.To}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}