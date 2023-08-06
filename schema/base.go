package schema

import (
	"embed"

	"github.com/PromptPal/PromptPal/service"
	"github.com/sirupsen/logrus"
)

var hashidService service.HashIDService

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
	"types/project.gql",
	"types/openToken.gql",
	"types/prompt.gql",
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
) {
	hashidService = hi
}