package config

import (
	"os"
	"strconv"
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
	// Create a new Config instance
	config := &Config{}

	// Retrieve and assign the values from environment variables
	config.BucketName = os.Getenv("BUCKET_NAME")
	config.Region = os.Getenv("REGION")
	config.DownloadURLTimeLimit, _ = strconv.Atoi(os.Getenv("DOWNLOAD_URL_TIME_LIMIT"))
	config.PaginationPageSize, _ = strconv.Atoi(os.Getenv("PAGINATION_PAGE_SIZE"))
	config.AwsAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	config.AwsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	return config, nil
}
