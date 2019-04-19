package parser_test

import (
	"bufio"
	"os"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/stretchr/testify/require"

	"github.com/dgraph-io/dgraph/gql"
	"github.com/dgraph-io/dgraph/parser/lexer"
	"github.com/dgraph-io/dgraph/parser/parser"
)

func TestGraphQuery(t *testing.T) {
	//input := []byte(`{friends_of_alice(func:eq(name)){haha}}`)
	/*

	 */
	input := []byte(`{
		 path as shortest(from: 0x1, to: 0x4) {
		  friend
		 }
		 path2 as shortest(from: 0x2, to: 0x3) {
		    friend
		 }
 pathQuery1(func: uid(path)) {
   name
 }
 pathQuery2(func: uid(path2)) {
   name
 }
}
`)

	lex := lexer.NewLexer(input)
	p := parser.NewParser()

	query, err := p.Parse(lex)
	if err != nil {
		panic(err)
	}
	qast, ok := query.(gql.Result)

	if !ok {
		t.Fatalf("Unable to convert to gql.Result")
	}

	gqlResult, err := gql.Parse(gql.Request{
		Str: string(input),
	})
	require.NoError(t, err, "the result from gql.Parse should suceeed")

	//spew.Dump(gqlResult)
	parserDumpF, err := os.OpenFile("parser.dump", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer parserDumpF.Close()
	parserWriter := bufio.NewWriter(parserDumpF)
	spew.Fdump(parserWriter, qast)
	parserWriter.Flush()

	gqlDumpF, err := os.OpenFile("gql.dump", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer gqlDumpF.Close()
	gqlWriter := bufio.NewWriter(gqlDumpF)
	spew.Fdump(gqlWriter, gqlResult)
	gqlWriter.Flush()

	require.True(t, reflect.DeepEqual(qast, gqlResult))
}
