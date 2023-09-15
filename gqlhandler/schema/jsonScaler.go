package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

// JSON GraphQl scaler type
var JSON = graphql.NewScalar(
	graphql.ScalarConfig{
		Name:         "JSON",
		Description:  "The `JSON` scaler type represents JSON values as specified by [ECMA-404](https://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf).",
		Serialize:    func(value interface{}) interface{} { return value },
		ParseValue:   func(value interface{}) interface{} { return value },
		ParseLiteral: parseJsonLiteral,
	},
)

func parseJsonLiteral(valueAST ast.Value) interface{} {
	kind := valueAST.GetKind()

	switch kind {
	case kinds.StringValue:
		return valueAST.GetValue()
	case kinds.BooleanValue:
		return valueAST.GetValue()
	case kinds.IntValue:
		return valueAST.GetValue()
	case kinds.FloatValue:
		return valueAST.GetValue()
	case kinds.ObjectValue:
		obj := make(map[string]interface{})
		for _, v := range valueAST.GetValue().([]*ast.ObjectField) {
			obj[v.Name.Value] = parseJsonLiteral(v.Value)
		}
		return obj
	case kinds.ListValue:
		list := make([]interface{}, 0)
		for _, v := range valueAST.GetValue().([]ast.Value) {
			list = append(list, parseJsonLiteral(v))
		}
		return list
	default:
		return nil
	}

}
