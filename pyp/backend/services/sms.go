package services

import (
	"errors"
	"log"
	"os"
)

func SendSMSVerifyCode(phone, code string) error {
	if os.Getenv("SMS_DEBUG") == "1" {
		log.Printf("SMS mock to %s: %s", phone, code)
		return nil
	}

	return errors.New("sms service not configured")
}
