package infra

import (
	"encoding/base64"
	"github.com/hasura/go-graphql-client"
	"net/http"
)

var GraphQL *graphql.Client

type GraphQLModel struct {
	Endpoint string
}

type IGraphQLConfig interface {
	Open()
	OpenWithBasicAuth(username string, password string)
}

func NewGraphQLConfig(model GraphQLModel) IGraphQLConfig {
	return GraphQLModel{
		Endpoint: model.Endpoint,
	}
}

func (g GraphQLModel) Open() {

	client := graphql.NewClient(g.Endpoint, http.DefaultClient)
	GraphQL = client
}

func (g GraphQLModel) OpenWithBasicAuth(username string, password string) {
	client := graphql.NewClient(g.Endpoint, http.DefaultClient)

	auth := username + ":" + password
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	client = client.WithRequestModifier(func(req *http.Request) {
		req.Header.Add("Authorization", basicAuth)
	})

	GraphQL = client
}
