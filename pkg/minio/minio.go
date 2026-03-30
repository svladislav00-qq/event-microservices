package minio

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client
var MinioBucket string
var MinioPublicURL string

func CreateCloud() {
	ctx := context.Background()

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKeyID := os.Getenv("MINIO_SECRET_KEY")
	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKeyID, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Connected to MinIO")

	MinioClient = client
	bucket := os.Getenv("MINIO_BUCKET")
	MinioBucket = bucket
	MinioPublicURL = os.Getenv("MINIO_PUBLIC_URL")

	location := os.Getenv("LOCATION")

	err = MinioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucket := MinioClient.BucketExists(ctx, bucket)
		if errBucket == nil && exists {
			fmt.Println("Bucket already exists:", bucket)
		} else {
			log.Fatalln(err)
		}
	} else {
		fmt.Println("Created bucket:", bucket)
	}
}
