package awsS3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// GetObjectByKey returnes the object specificed by the key from the specified
// bucketName using the provided s3Client.
//
// A successful retrieval of the document returns the io.ReadCloser (Body) and
// err == nil.
func GetObjectByKey(
	// ctx is the Context in which to run this GetObject call
	ctx context.Context,

	// s3Client is the s3.Client used
	s3Client *s3.Client,

	// bucketName is the name of the bucket to retrieve the key from
	bucketName string,

	// key is the full path to the object to return from the bucket
	key string,

	// timeout is the time to wait for the request to complete
	timeout time.Duration,
) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	inputObject := s3.GetObjectInput{
		Key:    &key,
		Bucket: &bucketName,
	}

	obj, err := s3Client.GetObject(ctx, &inputObject)
	if err != nil {
		return nil, fmt.Errorf("error retrieving the object '%s' from bucket '%s': %w", key, bucketName, err)
	}

	return obj.Body, nil
}


// const (
// 	MAX_UPLOAD_SIZE_MB uint32 = 25_000
// 	DEFAULT_LOG_LEVEL         = 2
// )

// func GetObjectByKey(
// 	ctx context.Context,
// 	s3Client *s3.Client,
// 	bucketName string,
// 	key string,
// ) (io.ReadCloser, error) {
// 	// fetchOwner := true
// 	// maxKeys := 1
// 	// inputObject := s3.GetObjectInput{
// 	// 	Bucket: &bucketName,
// 	// 	Key:    &key,
// 	// }

// 	inputObject := s3.GetObjectInput{
// 		Bucket: &bucketName,
// 	}

// 	// // object, err := s3Client.GetObject(ctx, &inputObject)
// 	// obj, err := s3Client.GetObject(ctx, &inputObject)
// 	// if err != nil {
// 	// 	return nil, fmt.Errorf("error getting object '%s' from bucket '%s': %w", key, bucketName, err)
// 	// }

// 	// return obj.ReadCloser, nil
// 	// // objReadCloser := object.ReadCloser
// 	// return nil, nil

// }
