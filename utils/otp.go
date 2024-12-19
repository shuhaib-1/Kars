package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// OTP represents a one-time password and its expiration time
type OTP struct {
	Code      string
	ExpiredAt time.Time
}

func InitFunc() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	log.Println("SENDGRID_API_KEY:", os.Getenv("SENDGRID_API_KEY"))

}

const OtpExpirationTime = time.Minute * 1

func GenerateOTP(email string, length int) (OTP, error) {
	if length <= 0 {
		return OTP{}, fmt.Errorf("length must be greater than 0")
	}

	digits := "0123456789"
	otp := make([]byte, length)

	for i := range otp {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return OTP{}, err
		}
		otp[i] = digits[index.Int64()]
	}

	OtpDetails := OTP{
		Code:      string(otp),
		ExpiredAt: time.Now().Add(OtpExpirationTime),
	}

	return OtpDetails, nil
}

func SendOtp(email, otp string) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	log.Println(apiKey)
	if apiKey == "" {
		return fmt.Errorf("SENDGRID_API_KEY environment variable not set")
	}

	from := mail.NewEmail("Kars", "shuhaibpa85@gmail.com")
	subject := "Your OTP Code"
	to := mail.NewEmail("User", email)
	plainTextContent := fmt.Sprintf("Your OTP code is: %s", otp)
	htmlContent := fmt.Sprintf("<strong>Your OTP code is: %s</strong>", otp)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode != 202 {
		return fmt.Errorf("failed to send email, status code: %d", response.StatusCode)
	}

	log.Printf("Email sent with status code: %d\n", response.StatusCode)
	return nil
}
