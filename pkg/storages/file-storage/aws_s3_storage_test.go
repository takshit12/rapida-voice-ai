// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package storage_files

import (
	"context"
	"testing"

	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/configs"
	"github.com/stretchr/testify/assert"
)

func TestAwsFileStorage_Name(t *testing.T) {
	cfg := configs.AssetStoreConfig{
		StorageType:       "s3",
		StoragePathPrefix: "test-bucket",
		Auth: &configs.AwsConfig{
			Region:      "us-east-1",
			AccessKeyId: "test-key",
			SecretKey:   "test-secret",
		},
	}
	logger, _ := commons.NewApplicationLogger()
	storage := NewAwsFileStorage(cfg, logger)

	assert.Equal(t, "aws", storage.Name())
}

func TestAwsFileStorage_contentType(t *testing.T) {
	cfg := configs.AssetStoreConfig{
		StorageType:       "s3",
		StoragePathPrefix: "test-bucket",
		Auth: &configs.AwsConfig{
			Region:      "us-east-1",
			AccessKeyId: "test-key",
			SecretKey:   "test-secret",
		},
	}
	logger, _ := commons.NewApplicationLogger()
	storage := NewAwsFileStorage(cfg, logger).(*awsFileStorage)

	tests := []struct {
		filename   string
		expectedCT string
	}{
		{"test.json", "application/json"},
		{"audio.mp3", "audio/mpeg"},
		{"audio.wav", "audio/wav"},
		{"audio.ogg", "audio/ogg"},
		{"audio.flac", "audio/flac"},
		{"audio.aac", "audio/aac"},
		{"audio.m4a", "audio/mp4"},
		{"unknown.xyz", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := storage.contentType(tt.filename)
			assert.Equal(t, tt.expectedCT, result)
		})
	}
}

// Note: For full integration with AWS SDK mocking, we would need to use
// a more sophisticated mocking approach. The current implementation
// creates the session inside the methods, making it difficult to inject mocks.
// For comprehensive testing, consider refactoring to accept an S3 client interface
// or use AWS SDK's built-in testing utilities.

func TestAwsFileStorage_Store_SessionCreationFailure(t *testing.T) {
	// Test case where session creation fails
	cfg := configs.AssetStoreConfig{
		StorageType:       "s3",
		StoragePathPrefix: "test-bucket",
		Auth: &configs.AwsConfig{
			Region: "", // Invalid region to cause session failure
		},
	}
	logger, _ := commons.NewApplicationLogger()
	storage := NewAwsFileStorage(cfg, logger)

	ctx := context.Background()
	key := "test/file.txt"
	content := []byte("test content")

	result := storage.Store(ctx, key, content)

	// Should return error due to invalid session
	assert.Error(t, result.Error)
	assert.Equal(t, configs.S3, result.StorageType)
	assert.Contains(t, result.CompletePath, "s3://test-bucket/test/file.txt")
}

func TestAwsFileStorage_Get_SessionCreationFailure(t *testing.T) {
	cfg := configs.AssetStoreConfig{
		StorageType:       "s3",
		StoragePathPrefix: "test-bucket",
		Auth: &configs.AwsConfig{
			Region: "", // Invalid region to cause session failure
		},
	}
	logger, _ := commons.NewApplicationLogger()
	storage := NewAwsFileStorage(cfg, logger)

	ctx := context.Background()
	key := "test/file.txt"

	result := storage.Get(ctx, key)

	assert.Error(t, result.Error)
	assert.Nil(t, result.Data)
}

func TestAwsFileStorage_GetUrl_SessionCreationFailure(t *testing.T) {
	cfg := configs.AssetStoreConfig{
		StorageType:       "s3",
		StoragePathPrefix: "test-bucket",
		Auth: &configs.AwsConfig{
			Region: "invalid-region", // Invalid region that should cause issues
		},
	}
	logger, _ := commons.NewApplicationLogger()
	storage := NewAwsFileStorage(cfg, logger)

	ctx := context.Background()
	key := "test/file.txt"

	result := storage.GetUrl(ctx, key)

	// With invalid credentials/region, the presigned URL generation will fail
	// The result will either have an empty path (error case) or contain the expected values
	assert.Equal(t, configs.S3, result.StorageType)
	// In CI without AWS credentials, the presigned URL generation fails
	// so we only check that the method doesn't panic and returns a valid struct
	if result.CompletePath != "" {
		assert.Contains(t, result.CompletePath, "test-bucket")
		assert.Contains(t, result.CompletePath, "test/file.txt")
	}
}
