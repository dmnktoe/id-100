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

// DeleteFromS3 extracts the file key from the image URL and deletes it from S3
func DeleteFromS3(imageURL string) error {
	// Extract the filename from the URL path
	// Example: /storage/v1/object/public/id100-images/derive_1_1234567890.webp
	// or: https://xxx.supabase.co/storage/v1/object/public/id100-images/derive_1_1234567890.webp
	var fileName string

	if strings.Contains(imageURL, "/storage/v1/object/public/") {
		parts := strings.Split(imageURL, "/storage/v1/object/public/")
		if len(parts) == 2 {
			// Extract bucket and file path
			pathParts := strings.SplitN(parts[1], "/", 2)
			if len(pathParts) == 2 {
				fileName = pathParts[1]
			}
		}
	}

	if fileName == "" {
		return fmt.Errorf("could not extract filename from URL: %s", imageURL)
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
