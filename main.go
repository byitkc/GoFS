package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"log/slog"

	"github.com/a-h/templ"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/byitkc/GoFS/handler"
	"github.com/byitkc/GoFS/pkg/awsS3"
	"github.com/byitkc/GoFS/pkg/config"
	"github.com/byitkc/GoFS/pkg/files"
	"github.com/byitkc/GoFS/view/home"
	"github.com/byitkc/GoFS/view/settings"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func getS3Client(conf config.Config) *s3.Client {
	const (
		retryAttempts = 5
		httpTimeout   = 5 * time.Second
	)

	s3ClientOptions := s3.Options{
		AppID:            "gofs",
		Credentials:      conf.AWSConfig.Credentials,
		Region:           conf.AWSConfig.Region,
		RetryMaxAttempts: retryAttempts,
		RetryMode:        aws.RetryModeAdaptive,
		HTTPClient:       &http.Client{},
	}

	return s3.New(s3ClientOptions)
}

func GetObject() {
	logger := slog.New(slog.Default().Handler())

	config, err := config.LoadConfig()
	if err != nil {
		logger.Error("error loading configuration: %w", err)
		os.Exit(1)
	}

	s3Client := getS3Client(config)

	key := "deskdirector/DeskDirectorPortal.msi"
	getObjectByKeyTimeout := 5 * time.Second
	object, err := awsS3.GetObjectByKey(
		context.TODO(),
		s3Client,
		config.UploadConfig.BucketName,
		key,
		getObjectByKeyTimeout,
	)
	if err != nil {
		logger.Error("error getting object by key '%s': %w", key, err)
		os.Exit(1)
	}

	fmt.Println(object)
}

func uploadFileToS3(s3Client *s3.Client, bucketName, filePath string) error {
	encrypt := true
	if err := files.CheckSourceFile(filePath); err != nil {
		return fmt.Errorf("error locating file at '%s': %w", filePath, err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening the file at '%s': %s", filePath, err)
	}
	defer file.Close()
	hash, err := files.ComputeFileSHA256ToBase64(filePath)
	if err != nil {
		return fmt.Errorf("error hashing the file at '%s': %w", filePath, err)
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

	objectOutput, err := s3Client.PutObject(context.TODO(), &objectInput)
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	fmt.Println(objectOutput)

	return nil
}

func createContainingFolder(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating the directory: %s: %s", dirName, err.Error())
	}
	return nil
}

//go:embed public/*
var publicFiles embed.FS
var protocol = "https"
var baseDomain = "www.testing.com"
var uploadDirName = "uploads"

func main() {
	handler.UploadDir = "uploads"
	handler.Protocol = "http"
	handler.Hostname = "localhost"
	handler.Port = 3000
	createContainingFolder(uploadDirName)

	publicFS, err := fs.Sub(publicFiles, "public")
	if err != nil {
		panic(fmt.Sprintf("error mapping public files to a embedded filesystem: %s", err.Error()))
	}

	fs := http.FileServer(http.Dir("./uploads"))

	router := chi.NewRouter()
	router.Handle("/public/*", http.StripPrefix("/public", http.FileServerFS(publicFS)))
	router.Handle("/uploads/*", http.StripPrefix("/uploads/", fs))
	router.Get("/upload", handler.MakeHandler(handler.HandleUploadIndex))
	router.Post("/upload", handler.MakeHandler(handler.HandleUploadIndexPost))
	router.Handle("/settings", templ.Handler(settings.Index()))
	router.Handle("/", templ.Handler(home.Index()))

	fmt.Println("Listening on :3000")
	http.ListenAndServe("localhost:3000", router)
}

// func GetObject(ctx context.Context, c *s3.Client, bucketName, key string) (io.Reader, error) {

// 	listInput := &s3.ListObjectsV2Input{
// 		Bucket: aws.String(bucketName),
// 		Prefix: aws.String(key),
// 	}

// 	objectInput := s3.GetObjectInput{
// 		Bucket: aws.String(bucketName),
// 		Key:    aws.String(key),
// 	}

// 	object, err := c.GetObject(ctx, &objectInput)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error retrieving object s3://%s/%s: %s", bucketName, key)
// 	}

// 	output, err := c.ListObjectsV2(context.TODO(), listInput)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error: %s", err.Error())
// 	}

// 	return nil, nil

// 	// TODO: Check that the returned object is good
// 	// TODO: Read the content of the object or return an io.Reader(?)

// 	// var minCount = int32(0)
// 	// if output.KeyCount == &minCount {
// 	// 	return nil, fmt.Errorf("no objects were returned for the search %s", prefix)
// 	// }
// 	// if output.KeyCount > aws.Int32(1) {

// 	// }
// }

// func dev() {
// 	ctx := context.Background()
// 	logger := config.InitLogger(0)

// 	logger.Info(fmt.Sprintf("absolute maximum upload size %d MB", config.MAX_UPLOAD_SIZE_MB))

// 	config, err := config.LoadConfig()
// 	if err != nil {
// 		slog.Error(err.Error())
// 	}

// 	client := s3.NewFromConfig(config.AWSConfig)

// 	credentials, err := config.AWSConfig.Credentials.Retrieve(ctx)
// 	if err != nil {
// 		slog.Error("error retreiving credentials", "error", err.Error())
// 		os.Exit(1)
// 	}

// 	err = awss3.GetObjectsFromBucketByName(ctx, client, credentials, config.UploadConfig.BucketName)
// 	if err != nil {
// 		// errString := fmt.Sprintf("unable to retreieve the objects in bucket: '%s': %w", config.UploadConfig.BucketName, err)
// 		slog.Error("unable to retreieve the objects in bucket:", "bucketName", config.UploadConfig.BucketName, "error", err)
// 		os.Exit(1)
// 	}

// filePath := "test.txt"

// if err := uploadFileToS3(client, config.UploadConfig.BucketName, filePath); err != nil {
// 	slog.Error("error uploading '%s' to S3: %w", filePath, err)
// }

// objectToGet := "deskdirector/DeskDirectorPortal.msi"
// object, err := GetObject(ctx, client, config.BucketName, objectToGet)
// if err != nil {
// 	glog.Fatalf("Unable to retrieve object s3://%s/%s", config.BucketName, objectToGet)
// }

// bufio.NewScanner(object)

// for scanner.Scan() {

// }

// bufio.Reader(object)

// listInput := &s3.ListObjectsV2Input{
// 	Bucket: aws.String(config.BucketName),
// }
// output, err := client.ListObjectsV2(context.TODO(), listInput)
// if err != nil {
// 	glog.Fatalf(err.Error())
// }

// glog.Flush()

// i := 1
// glog.Infoln("First page:")
// for _, object := range output.Contents {
// 	fmt.Printf("file=%d, key=%s, size=%d\n", i, aws.ToString(object.Key), object.Size)
// 	i++
// }

// }

// func uploadTestFile(client *s3.Client, path string) error {
// 	if !doesFileExist(path) {
// 		return fmt.Errorf("file doesn't exist at path '%s'", path)
// 	}
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return fmt.Errorf("error opening file: %s", err.Error())
// 	}
// 	defer file.Close()

// 	i := 1
// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		fmt.Println("Line %d: %s\n", i, scanner.Text())
// 		i++
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return fmt.Errorf("error reading file: %s", err.Error())
// 	}

// 	return nil
// }
