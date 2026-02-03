package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret                  string
	CookieName                 string
	MongoDBURL                 string
	PostgressURL               string
	RedisUrl                   string
	RedisUserName              string
	RedisPassword              string
	SMTPEmail                  string
	SMTPPassword               string
	FrontEndUrl                string
	EsewaMerchantCode          string
	EsewaSecretKey             string
	ESewaSuccessURL            string
	ESewaFailedURL             string
	ESewaPaymentURL            string
	EsewaPaymentStatusCheckURL string
}

var AppConfig *Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		log.Printf("Waring : Couldnot load .env files : %v", err)
		return fmt.Errorf("couldn't load .env files : %v", err)
	}
	AppConfig = &Config{
		JWTSecret:                  os.Getenv("JWT_SECRET"),
		CookieName:                 "user_token",
		MongoDBURL:                 os.Getenv("MONGO_URL"),
		PostgressURL:               os.Getenv("POSTGRESS_URL"),
		RedisUrl:                   os.Getenv("REDIS_URL"),
		RedisUserName:              os.Getenv("REDIS_USERNAME"),
		RedisPassword:              os.Getenv("REDIS_PASSWORD"),
		SMTPEmail:                  os.Getenv("SMTP_EMAIL"),
		SMTPPassword:               os.Getenv("SMTP_PASSWORD"),
		FrontEndUrl:                os.Getenv("FRONTEND_URL"),
		EsewaMerchantCode:          os.Getenv("ESEWA_MERCHANT_CODE"),
		EsewaSecretKey:             os.Getenv("ESEWA_SECRET_KEY"),
		ESewaPaymentURL:            os.Getenv("ESEWA_PAYMENT_URL"),
		EsewaPaymentStatusCheckURL: os.Getenv("ESEWA_PAYMENT_STATUS_CHECK_URL"),
	}
	AppConfig.ESewaSuccessURL = fmt.Sprintf("%s/products/checkout/payment/success", AppConfig.FrontEndUrl)
	AppConfig.ESewaFailedURL = fmt.Sprintf("%s/products/checkout/payment/failed", AppConfig.FrontEndUrl)
	return nil
}
