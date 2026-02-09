package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// extractFileNameFromURL extracts the filename from a MinIO/S3 storage URL
func extractFileNameFromURL(imageURL string) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("empty URL")
	}

	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "id100-images"
	}

	// Handle MinIO URL format: http://minio:9000/bucket-name/filename.ext
	// or: http://localhost:9000/bucket-name/filename.ext
	if strings.Contains(imageURL, "/"+bucket+"/") {
		parts := strings.Split(imageURL, "/"+bucket+"/")
		if len(parts) == 2 {
			return parts[1], nil
		}
	}

	// Handle relative path: bucket-name/filename.ext
	fileName := strings.TrimLeft(imageURL, "/")
	if strings.HasPrefix(fileName, bucket+"/") {
		return strings.TrimPrefix(fileName, bucket+"/"), nil
	}

	// If it's just a filename or nested path (no URL), return as-is
	// Examples: "derive_5_1.webp" or "subfolder/image.jpg"
	return fileName, nil
}

// DeleteFromS3 extracts the file key from the image URL and deletes it from S3/MinIO
func DeleteFromS3(imageURL string) error {
	// Extract the filename from the URL path
	// Example: http://minio:9000/id100-images/derive_1_1234567890.webp
	// or: http://localhost:9000/id100-images/derive_1_1234567890.webp
	fileName, err := extractFileNameFromURL(imageURL)
	if err != nil {
		return err
	}

	// Create S3 client
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("S3_REGION")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("S3_ACCESS_KEY"),
			os.Getenv("S3_SECRET_KEY"),
			""),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to load S3 config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	// Delete the object from S3
	_, err = s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(fileName),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	log.Printf("Successfully deleted %s from S3", fileName)
	return nil
}
