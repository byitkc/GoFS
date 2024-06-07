package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/golang/glog"
	"github.com/joho/godotenv"
)

const (
	MAX_UPLOAD_SIZE_MB uint32 = 25_000
	DEFAULT_LOG_LEVEL  int64  = 2
	MAX_LOG_LEVEL      int64  = 4
	MIN_LOG_LEVEL      int64  = 0
)

type Config struct {
	// AWSConfig stores the configuration for AWS
	AWSConfig aws.Config

	// UploadConfig stores the configuration for uploading documents to AWS
	UploadConfig UploadConfig
}

type UploadConfig struct {
	BucketName    string
	MaxUploadSize uint32
}

type AWSCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
}

type UploadFile struct {
	path        string
	hash        string
	contentType string
}

func (creds AWSCredentials) Retrieve(ctx context.Context) (aws.Credentials, error) {
	credentials := aws.Credentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
	}

	return credentials, nil
}

func InitLogger(level int) *slog.Logger {
	logHandlerOptions := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, logHandlerOptions))
}

func LoadConfig() (Config, error) {
	var (
		maxUploadSizeMb int64
	)

	if err := godotenv.Load(); err != nil {
		return Config{}, fmt.Errorf("failed to load environment variables from '.env', ensure that the file exists and is accessible: %s", err.Error())
	}

	awsRegion := os.Getenv("AWS_REGION")
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsBucketName := os.Getenv("AWS_BUCKET_NAME")
	maxUploadSize := os.Getenv("MAX_UPLOAD_SIZE_MB")

	if awsAccessKey == "" {
		return Config{}, fmt.Errorf("unable to load the AWS Access Key ID, please ensure that it's setup in your .env or environment variables as the 'AWS_ACCESS_KEY_ID")
	}
	if awsSecretKey == "" {
		return Config{}, fmt.Errorf("unable to load the AWS Secret Access Key, please ensure that it's setup in your .env or environment variables as the 'AWS_SECRET_ACCESS_KEY")
	}

	if maxUploadSize == "" {
		maxUploadSizeMb = 100
		glog.Warning("max upload size not specified, setting to %dMB", maxUploadSizeMb)
	} else {
		var err error
		maxUploadSizeMb, err = strconv.ParseInt(maxUploadSize, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("unable to determine the max upload size: %s", err.Error())
		}
	}
	if maxUploadSizeMb >= int64(^uint32(0)) {
		return Config{}, fmt.Errorf("max upload size exceeds the size limit (~420TB), please reduce the 'MAX_UPLOAD_SIZE' environment variable to not exceed 2^32-1")
	}

	if awsRegion == "" {
		glog.Warning("using 'us-east-1' for the region, as no region was set")
		awsRegion = "us-east-1"
	} else if len(strings.Split(awsRegion, "-")) < 3 {
		return Config{}, fmt.Errorf("region name appears to be invalid: %s, it should be in the format 'country-region-#', Ex. 'us-east-1'", awsRegion)
	}

	awsCredentials := AWSCredentials{
		AccessKeyID:     awsAccessKey,
		SecretAccessKey: awsSecretKey,
	}

	config := Config{
		AWSConfig: aws.Config{
			Region:      awsRegion,
			Credentials: awsCredentials,
		},
		UploadConfig: UploadConfig{
			BucketName:    awsBucketName,
			MaxUploadSize: uint32(maxUploadSizeMb),
		},
	}

	return config, nil
}
