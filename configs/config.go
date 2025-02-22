package configs

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/viper"
)

type Config struct {
	AWS_ACCESS_KEY_ID               string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWS_SECRET_ACCESS_KEY           string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWS_BUCKET_NAME                 string `mapstructure:"AWS_S3_BUCKET_NAME"`
	AWS_REGION                      string `mapstructure:"AWS_REGION"`
	AWS_S3_BUCKET_ACCESS_KEY        string `mapstructure:"AWS_S3_BUCKET_ACCESS_KEY"`
	AWS_S3_BUCKET_SECRET_ACCESS_KEY string `mapstructure:"AWS_S3_BUCKET_SECRET_ACCESS_KEY"`
	TIMEOUT                         int    `mapstructure:"TIMEOUT"`
	API_GROUP                       string `mapstructure:"API_GROUP"`
	PORT                            int    `mapstructure:"PORT"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(filepath.Join(path))
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	return config, nil
}

func ConnectAWS(cfg Config) (*s3.Client, *s3.PresignClient, error) {
	timeout := time.Duration(cfg.TIMEOUT) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.AWS_REGION),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWS_ACCESS_KEY_ID,
			cfg.AWS_SECRET_ACCESS_KEY,
			"",
		)),
	)
	if err != nil {
		return nil, nil, err
	}

	client := s3.NewFromConfig(awsCfg)
	presignClient := s3.NewPresignClient(client)

	// Check if bucket exists
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &cfg.AWS_BUCKET_NAME,
	})
	if err != nil {
		// If bucket doesn't exist, create it
		_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: &cfg.AWS_BUCKET_NAME,
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraint(cfg.AWS_REGION),
			},
		})
		if err != nil {
			log.Printf("Couldn't create bucket %v. Here's why: %v\n", cfg.AWS_BUCKET_NAME, err)
			return nil, nil, err
		}
		log.Printf("Created bucket %v in %v\n", cfg.AWS_BUCKET_NAME, cfg.AWS_REGION)
	}

	return client, presignClient, nil
}
