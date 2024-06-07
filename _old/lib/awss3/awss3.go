package awss3

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/byitkc/GoFS/pkg/files"
	"github.com/google/uuid"
)

const (
	MAX_UPLOAD_SIZE_MB uint32 = 25_000
	DEFAULT_LOG_LEVEL  int64  = 2
	MAX_LOG_LEVEL      int64  = 4
	MIN_LOG_LEVEL      int64  = 0
)

// func GetBucket(bucketName string) {

// }

func GetObjectsFromBucket(
	ctx context.Context,
	client *s3.Client,
	credentials aws.Credentials,
	bucketName string,
) error {
	slog.Info(fmt.Sprintf("retrieving objects from bucket: '%s'", bucketName))

	// listInput := &s3.ListObjectsV2Input{
	// 	Bucket: aws.String(bucketName),
	// }

	awsConfig := aws.Config{
		Region: "us-east-1",
	}

	fmt.Println(awsConfig)

	// client := s3.NewFromConfig()

	slog.Info("this function has yet to be implemented, returning error")
	return fmt.Errorf("the function GetObjectsFromBucketByName is not yet implemented")
}

func PutObjectInBucket(
	ctx context.Context,
	client *s3.Client,
	credentials aws.Credentials,
	bucketName string,
	filePath string,
) (string, error) {
	encrypt := true
	if err := files.CheckSourceFile(filePath); err != nil {
		return "", fmt.Errorf("error locating the file at '%s': %w", filePath, err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening the file at '%s': %w", filePath, err)
	}
	defer file.Close()

	hash, err := files.ComputeFileSHA256ToBase64(filePath)
	if err != nil {
		return "", fmt.Errorf("error hashing the file at '%s': %w", filePath, err)
	}

	uuid := uuid.New()
	key := fmt.Sprintf("%s/%s", uuid, filePath)

	objectInput := s3.PutObjectInput{
		Bucket:            &bucketName,
		Key:               &key,
		Body:              file,
		BucketKeyEnabled:  &encrypt,
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
		ACL:               types.ObjectCannedACLPublicRead,
		ChecksumSHA256:    &hash,
	}

	objectOutput, err := s3Client.PutObject(ctx, &objectInput)
	if err != nil {
		return "", fmt.Errorf("failed to upload file '%s' to S3: %w", filePath, err)

	}

	return "", nil
}
