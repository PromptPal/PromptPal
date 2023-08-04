package schema

import (
	"embed"

	"github.com/sirupsen/logrus"
)

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
