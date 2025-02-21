package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func GetAWSConfigRegion() (region string) {
	region = os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-southeast-1" // Default to us-east-1 if not provided
	}

	return region
}

func GetAWSConfig() (accessKey string, secretKey string, region string) {
	accessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	region = GetAWSConfigRegion()

	return accessKey, secretKey, region
}
