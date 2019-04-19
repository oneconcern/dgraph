package ast

import (
	"github.com/dgraph-io/dgraph/gql"
	"github.com/dgraph-io/dgraph/parser/token"
)

type Attrib interface{}

func NewGraphQuery(id Attrib) (*gql.GraphQuery, error) {
	return &gql.GraphQuery{
		Attr: string(id.(*token.Token).Lit),
	}, nil
}
