package ast

import (
	"github.com/dgraph-io/dgraph/gql"
	"github.com/dgraph-io/dgraph/parser/token"
)

type Attrib interface{}

func NewGraphQuery(queryList Attrib) (gql.Result, error) {
	queries := queryList.(QueryList)
	queryVars := getQueryVars(queries)
	return gql.Result{
		Query:     queries,
		QueryVars: queryVars,
	}, nil
}

func getQueryVars(queries QueryList) []*gql.Vars {
	vars := make([]*gql.Vars, 0)
	for _, query := range queries {
		qvar := &gql.Vars{
			Needs: varNames(query.NeedsVar),
		}
		if len(query.Var) > 0 {
			qvar.Defines = []string{query.Var}
		}
		vars = append(vars, qvar)
	}
	return vars
}

func varNames(vars []gql.VarContext) []string {
	varNames := make([]string, 0, len(vars))
	for _, v := range vars {
		varNames = append(varNames, v.Name)
	}
	return varNames
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
		NeedsVar: theFunc.needsVar,
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

	Name     string
	needsVar []gql.VarContext
}

type Function struct {
	argsForQuery map[string]string
	realFunc     *gql.Function
	children     []*gql.GraphQuery
	needsVar     []gql.VarContext
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
		needsVar:     theFuncHead.needsVar,
	}, nil
}

func NewFuncHead(name, funcHeadContent Attrib) (FuncHead, error) {
	funcHead := FuncHead{
		Name:     tokStr(name),
		needsVar: make([]gql.VarContext, 0),
	}
	if funcHead.Name == "val" {
		funcHead.needsVar = append(funcHead.needsVar, gql.VarContext{
			Name: tokStr(funcHeadContent),
			Typ:  gql.ValueVar,
		})
	} else if funcHead.Name == "uid" {
		funcHead.needsVar = append(funcHead.needsVar, gql.VarContext{
			Name: tokStr(funcHeadContent),
			Typ:  gql.UidVar,
		})
	}
	return funcHead, nil
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
