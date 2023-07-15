package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BucketName           string `json:"bucketName"`
	Region               string `json:"region"`
	DownloadURLTimeLimit int    `json:"downloadURLTimeLimit"`
	PaginationPageSize   int    `json:"paginationPageSize"`
	AwsAccessKeyID       string `json:"awsAccessKeyId"`
	AwsSecretAccessKey   string `json:"awsSecretAccessKey"`
}

func LoadConfig() (*Config, error) {
	// Load the environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return nil, err
	}

	// Create a new Config instance
	config := &Config{}

	// Retrieve and assign the values from environment variables
	config.BucketName = os.Getenv("BUCKET_NAME")
	config.Region = os.Getenv("REGION")

	// Parse integer values
	downloadURLTimeLimitStr := os.Getenv("DOWNLOAD_URL_TIME_LIMIT")
	config.DownloadURLTimeLimit, err = strconv.Atoi(downloadURLTimeLimitStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DOWNLOAD_URL_TIME_LIMIT: %v", err)
	}

	paginationPageSizeStr := os.Getenv("PAGINATION_PAGE_SIZE")
	config.PaginationPageSize, err = strconv.Atoi(paginationPageSizeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PAGINATION_PAGE_SIZE: %v", err)
	}

	config.AwsAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	config.AwsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	return config, nil
}
