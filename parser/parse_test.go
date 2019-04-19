package parser_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/dgraph-io/dgraph/gql"
	"github.com/dgraph-io/dgraph/parser/lexer"
	"github.com/dgraph-io/dgraph/parser/parser"
)

func TestGraphQuery(t *testing.T) {
	input := []byte(`{friends_of_alice(func:eq(name)){haha}}`)
	lex := lexer.NewLexer(input)
	p := parser.NewParser()

	query, err := p.Parse(lex)
	if err != nil {
		panic(err)
	}
	qast, ok := query.(*gql.GraphQuery)
	if !ok {
		t.Fatalf("Unable to convert to gql.GraphQuery")
	}
	spew.Dump(qast)
}
