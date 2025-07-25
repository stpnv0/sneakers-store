package repository

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cfg_pkg "product_service/internal/config"
)

type FileStoreRepository struct {
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
	log           *slog.Logger
}

func NewFileStoreRepository(ctx context.Context, config *cfg_pkg.Config, log *slog.Logger) (*FileStoreRepository, error) {
	const op = "repository.NewFileStoreRepository"
	log = log.With(slog.String("op", op))

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		endpoint := fmt.Sprintf("http://%s", config.S3.Endpoint)
		return aws.Endpoint{
			URL:           endpoint,
			SigningRegion: "us-east-1",
			Source:        aws.EndpointSourceCustom,
		}, nil
	})

	awsCfg, err := aws_config.LoadDefaultConfig(ctx,
		aws_config.WithEndpointResolverWithOptions(resolver),
		aws_config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.S3.AccessKey, config.S3.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	repo := &FileStoreRepository{
		s3Client:      s3Client,
		presignClient: s3.NewPresignClient(s3Client),
		bucketName:    "sneakers", // Можно вынести в конфиг
		log:           log,
	}

	err = repo.ensureBucket(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure bucket: %w", err)
	}

	log.Info("successfully connected to MinIO and ensured bucket exists", slog.String("bucket", repo.bucketName))
	return repo, nil
}

func (r *FileStoreRepository) ensureBucket(ctx context.Context) error {
	_, err := r.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucketName),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			r.log.Info("bucket not found, creating new one", slog.String("bucket", r.bucketName))
			_, createErr := r.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(r.bucketName),
			})
			if createErr != nil {
				return fmt.Errorf("failed to create bucket: %w", createErr)
			}
			policy := fmt.Sprintf(`{
                "Version": "2012-10-17",
                "Statement": [
                    {
                        "Effect": "Allow",
                        "Principal": "*",
                        "Action": ["s3:GetObject"],
                        "Resource": ["arn:aws:s3:::%s/*"]
                    }
                ]
            }`, r.bucketName)
			_, policyErr := r.s3Client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
				Bucket: aws.String(r.bucketName),
				Policy: aws.String(policy),
			})
			if policyErr != nil {
				return fmt.Errorf("failed to set bucket policy: %w", policyErr)
			}
			r.log.Info("bucket created and policy set successfully")
			return nil
		}
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	r.log.Info("bucket already exists")
	return nil
}

func (r *FileStoreRepository) GenerateUploadURL(ctx context.Context, key string, contentType string) (string, error) {
	req, err := r.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}
	return req.URL, nil
}
