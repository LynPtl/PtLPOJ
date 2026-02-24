package auth

import (
	"fmt"
	"log"
)

// SendOTPEmail is a mock implementation of an email sender.
// In a real production deployment, this would be replaced by an SMTP or SendGrid/AWS SES client.
func SendOTPEmail(email, code string) error {
	// [TODO / SMTP SLOT]: Replace this block with real SMTP dialing logic later
	// e.g. using net/smtp or a 3rd party API like SendGrid

	line := "===============================================\n"
	line += fmt.Sprintf("📧 MOCK EMAIL DISPATCHER 📧\n")
	line += fmt.Sprintf("To: %s\n", email)
	line += fmt.Sprintf("Subject: Your PtLPOJ Login Verification Code\n")
	line += "\n"
	line += fmt.Sprintf("Your OTP is: [%s]\n", code)
	line += "This code will expire in 5 minutes.\n"
	line += "==============================================="

	log.Println("\n" + line)
	return nil
}
