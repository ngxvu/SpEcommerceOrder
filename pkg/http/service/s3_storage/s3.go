package s3_storage

import (
	"basesource/conf"
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"mime/multipart"
	"net/http"
)

type S3Uploader struct {
	Client     *s3.S3
	BucketName string
	Region     string
}

func NewS3Uploader() (*S3Uploader, error) {
	config := conf.GetConfig()

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(config.S3Region),
		Endpoint:         aws.String(config.S3Endpoint),
		Credentials:      credentials.NewStaticCredentials(config.S3AccessKey, config.S3SecretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return &S3Uploader{
		Client:     s3.New(sess),
		BucketName: config.S3Bucket,
		Region:     config.S3Region,
	}, nil
}

func (u *S3Uploader) UploadFile(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	defer file.Close()

	// Read file content
	size := fileHeader.Size
	buffer := make([]byte, size)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	fileName := fileHeader.Filename
	contentType := http.DetectContentType(buffer)

	_, err = u.Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(u.BucketName),
		Key:           aws.String(fileName),
		ACL:           aws.String("public-read"),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return "", err
	}

	// Return public URL
	url := fmt.Sprintf("https://%s.%s.digitaloceanspaces.com/%s", u.BucketName, u.Region, fileName)
	return url, nil
}
