package configs

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	ACCESS_KEY_ID               string `mapstructure:"ACCESS_KEY_ID"`
	SECRET_ACCESS_KEY           string `mapstructure:"SECRET_ACCESS_KEY"`
	BUCKET_NAME                 string `mapstructure:"S3_BUCKET_NAME"`
	REGION                      string `mapstructure:"REGION"`
	S3_BUCKET_ACCESS_KEY        string `mapstructure:"S3_BUCKET_ACCESS_KEY"`
	S3_BUCKET_SECRET_ACCESS_KEY string `mapstructure:"S3_BUCKET_SECRET_ACCESS_KEY"`
	TIMEOUT                     int    `mapstructure:"TIMEOUT"`
	API_GROUP                   string `mapstructure:"API_GROUP"`
	PORT                        int    `mapstructure:"PORT"`
}

func LoadConfig(path string) (config Config, err error) {
	// viper.AddConfigPath(filepath.Join(path))
	// viper.SetConfigName(".env")
	// viper.SetConfigType("env")

	// if err := viper.ReadInConfig(); err != nil {
	// 	log.Fatalf("Error reading config file, %s", err)
	// }

	// err = viper.Unmarshal(&config)
	// if err != nil {
	// 	log.Fatalf("Unable to decode into struct, %v", err)
	// }

	timeout, _ := strconv.Atoi(os.Getenv("TIMEOUT"))
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	config = Config{
		ACCESS_KEY_ID:               os.Getenv("ACCESS_KEY_ID"),
		SECRET_ACCESS_KEY:           os.Getenv("SECRET_ACCESS_KEY"),
		BUCKET_NAME:                 os.Getenv("S3_BUCKET_NAME"),
		REGION:                      os.Getenv("REGION"),
		S3_BUCKET_ACCESS_KEY:        os.Getenv("S3_BUCKET_ACCESS_KEY"),
		S3_BUCKET_SECRET_ACCESS_KEY: os.Getenv("S3_BUCKET_SECRET_ACCESS_KEY"),
		TIMEOUT:                     timeout,
		API_GROUP:                   os.Getenv("API_GROUP"),
		PORT:                        port,
	}

	return config, nil
}

func ConnectAWS(cfg Config) (*s3.Client, *s3.PresignClient, error) {
	timeout := time.Duration(cfg.TIMEOUT) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.REGION),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.ACCESS_KEY_ID,
			cfg.SECRET_ACCESS_KEY,
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
		Bucket: &cfg.BUCKET_NAME,
	})
	if err != nil {
		// If bucket doesn't exist, create it
		_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: &cfg.BUCKET_NAME,
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraint(cfg.REGION),
			},
		})
		if err != nil {
			log.Printf("Couldn't create bucket %v. Here's why: %v\n", cfg.BUCKET_NAME, err)
			return nil, nil, err
		}
		log.Printf("Created bucket %v in %v\n", cfg.BUCKET_NAME, cfg.REGION)
	}

	return client, presignClient, nil
}
