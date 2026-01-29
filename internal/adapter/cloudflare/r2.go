package cloudflare

import (
	"bwanews/config"
	"bwanews/internal/core/domain/entity"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2/log"
)

var code string
var err error

type CloudflareR2Adapter interface {
	UploadImage(req *entity.FileUploadEntity) (string, error)
}

type cloudflareR2Adapter struct {
	Client  *s3.Client
	Bucket  string
	BaseUrl string
}

// UploadImage implements CloudflareR2Adapter.
func (c *cloudflareR2Adapter) UploadImage(req *entity.FileUploadEntity) (string, error) {
	openedFile, err := os.Open(req.Path)
	if err != nil {
		code = "[CLOUDFLARE R2] UploadImage = 1"
		log.Errorw(code, err)
		return "", err
	}

	defer openedFile.Close()

	_, err = c.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(c.Bucket),
		Key:         aws.String(req.Name),
		Body:        openedFile,
		ContentType: aws.String("image/jpeg"),
	})

	if err != nil {
		code = "[CLOUDFLARE R2] UploadImage = 2"
		log.Errorw(code, err)
		return "", err
	}

	return fmt.Sprintf("%s/%s", c.BaseUrl, req.Name), nil
}

func NewCloudflareR2Adapter(client *s3.Client, cfg *config.Config) CloudflareR2Adapter {
	clientBase := s3.NewFromConfig(cfg.LoadAwsConfig(), func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2.AccountID))
	})
	return &cloudflareR2Adapter{
		Client:  clientBase,
		Bucket:  cfg.R2.Name,
		BaseUrl: cfg.R2.PublicURL,
	}
}
