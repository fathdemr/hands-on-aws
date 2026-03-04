package test

import (
	"context"
	"fmt"
	"hands-on-aws/config"
	"hands-on-aws/services/S3Service"
	"testing"
)

func TestS3Service(t *testing.T) {
	var err error
	err = config.InitConfigFile("../../../")
	if err != nil {
		fmt.Println("InitConfigFile err:", err)
	}
	ctx := context.Background()
	service := S3Service.NewForPublicBucket(ctx)

	_, err = service.Upload("test/test.txt", ".txt", []byte("Hello, S3!"))
	if err != nil {
		fmt.Println("Error uploading file:", err)
	}
}
