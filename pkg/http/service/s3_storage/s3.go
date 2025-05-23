package s3_storage

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
	"io"
	"mime/multipart"
	"net/http"
)

type InfoS3 struct {
	S3SecretKey  string
	S3AccessKey  string
	S3BucketName string
	S3Region     string
	S3Endpoint   string
}

func GetS3Config(config *viper.Viper) (*InfoS3, error) {
	s3SecretKey := config.GetString("DigitalOcean.S3_SecretKey")
	s3AccessKey := config.GetString("DigitalOcean.S3_AccessKey")
	s3BucketName := config.GetString("DigitalOcean.S3_BucketName")
	s3Region := config.GetString("DigitalOcean.S3_Region")
	s3Endpoint := config.GetString("DigitalOcean.S3_Endpoint")

	return &InfoS3{
		S3SecretKey:  s3SecretKey,
		S3AccessKey:  s3AccessKey,
		S3BucketName: s3BucketName,
		S3Region:     s3Region,
		S3Endpoint:   s3Endpoint,
	}, nil
}

type S3Uploader struct {
	Client     *s3.S3
	BucketName string
	Region     string
}

func NewS3Uploader(info *InfoS3) (*S3Uploader, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(info.S3Region),
		Endpoint:         aws.String(info.S3Endpoint),
		Credentials:      credentials.NewStaticCredentials(info.S3AccessKey, info.S3SecretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return &S3Uploader{
		Client:     s3.New(sess),
		BucketName: info.S3BucketName,
		Region:     info.S3Region,
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
