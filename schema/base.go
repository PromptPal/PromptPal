package schema

import (
	"embed"

	"github.com/PromptPal/PromptPal/service"
	"github.com/sirupsen/logrus"
)

var web3Service service.Web3Service
var hashidService service.HashIDService
var rbacService *service.RBACService

type paginationInput struct {
	Limit  int32
	Offset int32
}

type QueryResolver struct{}

//go:embed schema.gql types/*.gql
var graphqlSchema embed.FS

var fileNames = []string{
	"schema.gql",
	"types/common.gql",
	"types/user.gql",
	"types/call.gql",
	"types/project.gql",
	"types/openToken.gql",
	"types/prompt.gql",
	"types/history.gql",
	"types/provider.gql",
}

func String() string {
	files := make([]byte, 0)

	for _, fn := range fileNames {
		b, err := graphqlSchema.ReadFile(fn)
		if err != nil {
			logrus.Panicln(err)
		}
		files = append(files, b...)
	}

	return string(files)
}

func Setup(
	hi service.HashIDService,
	w3 service.Web3Service,
	rbac *service.RBACService,
) {
	hashidService = hi
	web3Service = w3
	rbacService = rbac
}
