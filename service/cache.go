package service

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/cache/v9"
	"github.com/mitchellh/hashstructure/v2"
)

func Hash(s any) []byte {
	var b bytes.Buffer
	gob.NewEncoder(&b).Encode(s)
	return b.Bytes()
}

func generatePromptResponseCacheKey(promptID string, variables any) (string, error) {
	hash, err := hashstructure.Hash(variables, hashstructure.FormatV2, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("prompt-response:%s:%d", promptID, hash), nil
}

func SetPromptResponseCache(promptID string, variables any, value APIRunPromptResponse) error {
	k, err := generatePromptResponseCacheKey(promptID, variables)
	if err != nil {
		return err
	}
	Cache.Set(&cache.Item{
		Ctx:   context.Background(),
		Key:   k,
		Value: value,
		TTL:   time.Minute * 5,
	})
	return nil
}

func GetPromptResponseCache(promptID string, variables any) (*APIRunPromptResponse, bool, error) {
	k, err := generatePromptResponseCacheKey(promptID, variables)
	if err != nil {
		return nil, false, err
	}
	var result APIRunPromptResponse
	err = Cache.Get(context.Background(), k, &result)
	if err != nil {
		if !errors.Is(err, cache.ErrCacheMiss) {
			return nil, false, err
		}
		return nil, false, nil
	}
	return &result, true, nil
}
