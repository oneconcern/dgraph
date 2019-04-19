package ast

import (
	"github.com/dgraph-io/dgraph/gql"
	"github.com/dgraph-io/dgraph/parser/token"
)

type Attrib interface{}

func NewGraphQuery(queryList Attrib) (gql.Result, error) {
	return gql.Result{
		Query: queryList.(QueryList),
	}, nil
}

type QueryList []*gql.GraphQuery

func NewQueryList(query Attrib) (QueryList, error) {
	return QueryList{query.(*gql.GraphQuery)}, nil
}

func AppendQuery(queryList, query Attrib) (QueryList, error) {
	return append(queryList.(QueryList), query.(*gql.GraphQuery)), nil
}

func NewQuery(alias, function Attrib) (*gql.GraphQuery, error) {

	theFunc := function.(Function)

	return &gql.GraphQuery{
		Alias:    tokStr(alias),
		Args:     theFunc.argsForQuery,
		Func:     theFunc.realFunc,
		Children: theFunc.children,
	}, nil
}

func NewQueryWithVar(varName, alias, function Attrib) (*gql.GraphQuery, error) {
	gql, err := NewQuery(alias, function)
	gql.Var = tokStr(varName)
	return gql, err
}

func tokStr(tok Attrib) string {
	return string(tok.(*token.Token).Lit)
}

type FuncHead struct {
	Args         []gql.Arg
	ArgsForQuery map[string]string

	Name string
}

type Function struct {
	argsForQuery map[string]string
	realFunc     *gql.Function
	children     []*gql.GraphQuery
}

func NewFunction(funcHead, funcBody Attrib) (Function, error) {
	theFuncHead := funcHead.(FuncHead)
	theFuncBody := funcBody.(FuncBody)

	gqlFunc := &gql.Function{
		Args: theFuncHead.Args,
	}
	return Function{
		argsForQuery: theFuncHead.ArgsForQuery,
		realFunc:     gqlFunc,
		children:     theFuncBody.Children,
	}, nil
}

func NewFuncHead(name, funcHeadContent Attrib) (FuncHead, error) {
	return FuncHead{
		Name: tokStr(name),
	}, nil
}

func NewShortestPathFuncHead(from, to, optArgs Attrib) (FuncHead, error) {
	argsMap := make(map[string]string)
	argsMap["from"] = tokStr(from)
	argsMap["to"] = tokStr(to)
	return FuncHead{ArgsForQuery: argsMap}, nil
}

type FuncBody struct {
	Children []*gql.GraphQuery
}

func NewFuncBody(innerQueryList Attrib) (FuncBody, error) {
	return FuncBody{
		Children: innerQueryList.([]*gql.GraphQuery),
	}, nil
}

type InnerQuery *gql.GraphQuery
type InnerQueryList []*gql.GraphQuery

func NewInnerQueryList(innerQuery Attrib) ([]*gql.GraphQuery, error) {
	return InnerQueryList{
		&gql.GraphQuery{
			Attr: tokStr(innerQuery),
		},
	}, nil
}

func AppendInnerQuery(innerQueryList, innerQuery Attrib) (InnerQueryList, error) {
	return append(innerQueryList.(InnerQueryList), innerQuery.(InnerQuery)), nil
}
