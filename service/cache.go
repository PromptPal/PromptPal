package service

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/policy/lru"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/mitchellh/hashstructure/v2"
)

var apiPromptResponseCache *cache.Cache[string, APIRunPromptResponse] = cache.New[string, APIRunPromptResponse](
	cache.AsLRU[string, APIRunPromptResponse](
		lru.WithCapacity(1 << 10),
	),
)
var ApiPromptCache *cache.Cache[string, ent.Prompt]
var ProjectCache *cache.Cache[int, ent.Project]
var ProviderCache *cache.Cache[int, ent.Provider]
var PublicAPIAuthCache *cache.Cache[string, ent.OpenToken] = cache.New[string, ent.OpenToken]()

func init() {
	ApiPromptCache = cache.New[string, ent.Prompt]()
	ProjectCache = cache.New[int, ent.Project]()
	ProviderCache = cache.New[int, ent.Provider]()
}

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
	apiPromptResponseCache.Set(k, value, cache.WithExpiration(time.Minute*5))
	return nil
}

func GetPromptResponseCache(promptID string, variables any) (*APIRunPromptResponse, bool, error) {
	k, err := generatePromptResponseCacheKey(promptID, variables)
	if err != nil {
		return nil, false, err
	}
	result, ok := apiPromptResponseCache.Get(k)
	return &result, ok, nil
}
