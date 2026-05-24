package migrate

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const entryFieldsMigrationDescription = "contentful_entry.fields from map(json) to map(map(json))"

var (
	errInvalidExpressionRange = errors.New("invalid expression range")
	errLexExpression          = errors.New("lex expression")
	errParseConfig            = errors.New("parse config")
	errParseFieldsExpression  = errors.New("parse contentful_entry.fields expression")
)

type entryFieldsMigration struct{}

func (m entryFieldsMigration) Apply(path string, src []byte) ([]byte, fileReport, error) {
	file, diags := hclwrite.ParseConfig(src, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fileReport{}, fmt.Errorf("%w: %s: %s", errParseConfig, path, diags.Error())
	}

	report := fileReport{}

	for _, block := range file.Body().Blocks() {
		change, skip, migrated, err := migrateEntryBlock(path, block)
		if err != nil {
			return nil, fileReport{}, err
		}

		if skip != nil {
			report.Skips = append(report.Skips, *skip)
		}

		if migrated {
			report.Changes = append(report.Changes, change)
		}
	}

	if len(report.Changes) == 0 {
		return src, report, nil
	}

	return hclwrite.Format(file.Bytes()), report, nil
}

func migrateEntryBlock(path string, block *hclwrite.Block) (Change, *Skip, bool, error) {
	if block.Type() != "resource" {
		return Change{}, nil, false, nil
	}

	labels := block.Labels()
	if len(labels) != 2 || labels[0] != "contentful_entry" {
		return Change{}, nil, false, nil
	}

	address := fmt.Sprintf("%s.%s", labels[0], labels[1])

	fields := block.Body().GetAttribute("fields")
	if fields == nil {
		return Change{}, nil, false, nil
	}

	fieldTokens, migrated, skip, err := migrateEntryFieldsAttribute(path, address, fields)
	if err != nil {
		return Change{}, nil, false, err
	}

	if skip != nil {
		return Change{}, skip, false, nil
	}

	if !migrated {
		return Change{}, nil, false, nil
	}

	block.Body().SetAttributeRaw("fields", fieldTokens)

	return Change{
		Path:        path,
		Description: fmt.Sprintf("%s on %s", entryFieldsMigrationDescription, address),
	}, nil, true, nil
}

func migrateEntryFieldsAttribute(path string, address string, fields *hclwrite.Attribute) (hclwrite.Tokens, bool, *Skip, error) {
	exprBytes := fields.Expr().BuildTokens(nil).Bytes()

	expr, diags := hclsyntax.ParseExpression(exprBytes, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, false, nil, fmt.Errorf("%w: %s %s: %s", errParseFieldsExpression, path, address, diags.Error())
	}

	fieldsObject, ok := expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil, false, &Skip{
			Path:    path,
			Address: address,
			Reason:  "fields is not a literal object expression",
		}, nil
	}

	fieldAttrs := make([]hclwrite.ObjectAttrTokens, 0, len(fieldsObject.Items))
	changed := false

	for _, field := range fieldsObject.Items {
		fieldAttr, fieldChanged, skip, err := migrateEntryField(path, address, exprBytes, field)
		if err != nil {
			return nil, false, nil, err
		}

		if skip != nil {
			return nil, false, skip, nil
		}

		fieldAttrs = append(fieldAttrs, fieldAttr)
		changed = changed || fieldChanged
	}

	if !changed {
		return nil, false, nil, nil
	}

	return hclwrite.TokensForObject(fieldAttrs), true, nil, nil
}

func migrateEntryField(
	path string,
	address string,
	exprBytes []byte,
	field hclsyntax.ObjectConsItem,
) (hclwrite.ObjectAttrTokens, bool, *Skip, error) {
	fieldKey, keyStatic := staticObjectKey(field.KeyExpr)
	if !keyStatic {
		return hclwrite.ObjectAttrTokens{}, false, &Skip{
			Path:    path,
			Address: address,
			Reason:  "fields contains a dynamic field key",
		}, nil
	}

	fieldName, err := expressionTokens(exprBytes, field.KeyExpr)
	if err != nil {
		return hclwrite.ObjectAttrTokens{}, false, nil, fmt.Errorf("read %s %s fields.%s key: %w", path, address, fieldKey, err)
	}

	fieldObject, alreadyMigrated := field.ValueExpr.(*hclsyntax.ObjectConsExpr)
	if alreadyMigrated {
		fieldValue, err := expressionTokens(exprBytes, fieldObject)
		if err != nil {
			return hclwrite.ObjectAttrTokens{}, false, nil, fmt.Errorf("read %s %s fields.%s value: %w", path, address, fieldKey, err)
		}

		return hclwrite.ObjectAttrTokens{
			Name:  fieldName,
			Value: fieldValue,
		}, false, nil, nil
	}

	localesObject, skip := jsonEncodedLocaleObject(path, address, fieldKey, field.ValueExpr)
	if skip != nil {
		return hclwrite.ObjectAttrTokens{}, false, skip, nil
	}

	localeAttrs, skip, err := migrateEntryFieldLocales(path, address, fieldKey, exprBytes, localesObject)
	if err != nil {
		return hclwrite.ObjectAttrTokens{}, false, nil, err
	}

	if skip != nil {
		return hclwrite.ObjectAttrTokens{}, false, skip, nil
	}

	return hclwrite.ObjectAttrTokens{
		Name:  fieldName,
		Value: hclwrite.TokensForObject(localeAttrs),
	}, true, nil, nil
}

func jsonEncodedLocaleObject(path string, address string, fieldKey string, value hclsyntax.Expression) (*hclsyntax.ObjectConsExpr, *Skip) {
	call, isFunctionCall := value.(*hclsyntax.FunctionCallExpr)
	if !isFunctionCall || call.Name != "jsonencode" || call.ExpandFinal || len(call.Args) != 1 {
		return nil, &Skip{
			Path:    path,
			Address: address,
			Reason:  fmt.Sprintf("fields.%s is not a direct jsonencode({...}) expression", fieldKey),
		}
	}

	localesObject, ok := call.Args[0].(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil, &Skip{
			Path:    path,
			Address: address,
			Reason:  fmt.Sprintf("fields.%s jsonencode argument is not a literal locale object", fieldKey),
		}
	}

	return localesObject, nil
}

func migrateEntryFieldLocales(
	path string,
	address string,
	fieldKey string,
	exprBytes []byte,
	localesObject *hclsyntax.ObjectConsExpr,
) ([]hclwrite.ObjectAttrTokens, *Skip, error) {
	localeAttrs := make([]hclwrite.ObjectAttrTokens, 0, len(localesObject.Items))

	for _, locale := range localesObject.Items {
		localeKey, keyStatic := staticObjectKey(locale.KeyExpr)
		if !keyStatic {
			return nil, &Skip{
				Path:    path,
				Address: address,
				Reason:  fmt.Sprintf("fields.%s contains a dynamic locale key", fieldKey),
			}, nil
		}

		localeName, err := expressionTokens(exprBytes, locale.KeyExpr)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s %s fields.%s.%s key: %w", path, address, fieldKey, localeKey, err)
		}

		localeValue, err := expressionTokens(exprBytes, locale.ValueExpr)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s %s fields.%s.%s value: %w", path, address, fieldKey, localeKey, err)
		}

		localeAttrs = append(localeAttrs, hclwrite.ObjectAttrTokens{
			Name:  localeName,
			Value: hclwrite.TokensForFunctionCall("jsonencode", localeValue),
		})
	}

	return localeAttrs, nil, nil
}

func staticObjectKey(expr hclsyntax.Expression) (string, bool) {
	value, diags := expr.Value(nil)
	if diags.HasErrors() || !value.IsKnown() || value.Type() != cty.String {
		return "", false
	}

	return value.AsString(), true
}

func expressionTokens(src []byte, expr hclsyntax.Expression) (hclwrite.Tokens, error) {
	exprRange := expr.Range()
	if exprRange.Start.Byte < 0 || exprRange.End.Byte > len(src) || exprRange.Start.Byte > exprRange.End.Byte {
		return nil, fmt.Errorf("%w: %d:%d", errInvalidExpressionRange, exprRange.Start.Byte, exprRange.End.Byte)
	}

	tokens, diags := hclsyntax.LexExpression(src[exprRange.Start.Byte:exprRange.End.Byte], exprRange.Filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("%w: %s", errLexExpression, diags.Error())
	}

	return hclwriteTokens(tokens), nil
}

func hclwriteTokens(tokens hclsyntax.Tokens) hclwrite.Tokens {
	result := make(hclwrite.Tokens, 0, len(tokens))

	for _, token := range tokens {
		if token.Type == hclsyntax.TokenEOF {
			continue
		}

		result = append(result, &hclwrite.Token{
			Type:  token.Type,
			Bytes: token.Bytes,
		})
	}

	return result
}
