package service

import (
	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent"
)

var ApiPromptCache *cache.Cache[string, ent.Prompt]
var ProjectCache *cache.Cache[int, ent.Project]
var PublicAPIAuthCache *cache.Cache[string, int] = cache.New[string, int]()

func init() {
	ApiPromptCache = cache.New[string, ent.Prompt]()
	ProjectCache = cache.New[int, ent.Project]()
}
