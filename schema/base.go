package schema

import (
	"embed"

	"github.com/PromptPal/PromptPal/service"
	"github.com/sirupsen/logrus"
)

var hashidService service.HashIDService

type QueryResolver struct{}

//go:embed schema.gql types/*.gql
var graphqlSchema embed.FS

var fileNames = []string{
	"schema.gql",
	"types/project.gql",
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
