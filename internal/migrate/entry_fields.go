package migrate

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

var errParseConfig = errors.New("parse config")

// Edits use offsets from the original parse. Only the calls that encode field
// or locale values change; comments, expressions and unrelated bytes survive.
type sourceEdit struct {
	start, end int
	text       string
}

func migrateEntryFields(path string, src []byte) ([]byte, Report, error) {
	file, diags := hclsyntax.ParseConfig(src, path, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, Report{}, fmt.Errorf("%w: %w", errParseConfig, diags)
	}

	body, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "resource", LabelNames: []string{"type", "name"}}},
	})
	if diags.HasErrors() {
		return nil, Report{}, fmt.Errorf("%w: %w", errParseConfig, diags)
	}

	// HCL's expression range ends immediately after a heredoc marker. The
	// marker must finish its line before we can close a surrounding call.
	heredocEnds := make(map[int]string)

	tokens, _ := hclsyntax.LexConfig(src, path, hcl.InitialPos)
	for i, token := range tokens {
		if token.Type == hclsyntax.TokenCHeredoc && tokens[i+1].Type == hclsyntax.TokenNewline {
			heredocEnds[token.Range.End.Byte] = string(tokens[i+1].Bytes)
		}
	}

	var (
		report Report
		edits  []sourceEdit
	)

	for _, block := range body.Blocks {
		if block.Labels[0] != "contentful_entry" {
			continue
		}

		content, _, diags := block.Body.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{{Name: "fields"}},
		})
		if diags.HasErrors() {
			return nil, Report{}, fmt.Errorf("%w: %w", errParseConfig, diags)
		}

		fields := content.Attributes["fields"]
		if fields == nil {
			continue
		}

		address := "contentful_entry." + block.Labels[1]

		changes, reason := entryFieldEdits(src, fields.Expr, heredocEnds)
		if reason != "" {
			report.Skips = append(report.Skips, Skip{Path: path, Address: address, Reason: reason})
		} else if len(changes) > 0 {
			edits = append(edits, changes...)
			report.Changes = append(report.Changes, Change{Path: path, Description: "localize fields on " + address})
		}
	}

	if len(edits) == 0 {
		return src, report, nil
	}

	var (
		result bytes.Buffer
		pos    int
	)

	for _, edit := range edits {
		result.Write(src[pos:edit.start])
		result.WriteString(edit.text)
		pos = edit.end
	}

	result.Write(src[pos:])

	return result.Bytes(), report, nil
}

func entryFieldEdits(src []byte, fields hcl.Expression, heredocEnds map[int]string) ([]sourceEdit, string) {
	object, ok := fields.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil, "fields is not a literal object expression"
	}

	var edits []sourceEdit

	for _, field := range object.Items {
		name, static := staticObjectKey(field.KeyExpr)
		if !static {
			return nil, "fields contains a dynamic field key"
		}

		value := field.ValueExpr
		for {
			group, ok := value.(*hclsyntax.ParenthesesExpr)
			if !ok {
				break
			}

			value = group.Expression
		}

		if _, migrated := value.(*hclsyntax.ObjectConsExpr); migrated {
			continue
		}

		call, ok := value.(*hclsyntax.FunctionCallExpr)
		if !ok || call.Name != "jsonencode" || call.ExpandFinal || len(call.Args) != 1 {
			return nil, fmt.Sprintf("fields.%s is not a direct jsonencode({...}) expression", name)
		}

		locales, ok := call.Args[0].(*hclsyntax.ObjectConsExpr)
		if !ok {
			return nil, fmt.Sprintf("fields.%s jsonencode argument is not a literal locale object", name)
		}

		changes, reason := localizeJSONCall(src, call, locales, heredocEnds)
		if reason != "" {
			return nil, fmt.Sprintf("fields.%s %s", name, reason)
		}

		edits = append(edits, changes...)
	}

	return edits, ""
}

func localizeJSONCall(src []byte, call *hclsyntax.FunctionCallExpr, locales *hclsyntax.ObjectConsExpr, heredocEnds map[int]string) ([]sourceEdit, string) {
	// Retain the grouping parentheses when they enclose comments or line
	// breaks. Removing them could change expression boundaries.
	start, end := call.Range().Start.Byte, call.Range().End.Byte
	compact := bytes.Equal(src[start:locales.Range().Start.Byte], []byte("jsonencode(")) &&
		bytes.Equal(src[locales.Range().End.Byte:end], []byte(")"))

	prefixEnd := call.NameRange.End.Byte
	if compact {
		prefixEnd = locales.Range().Start.Byte
	}

	edits := []sourceEdit{{start: start, end: prefixEnd}}

	for _, locale := range locales.Items {
		if _, static := staticObjectKey(locale.KeyExpr); !static {
			return nil, "contains a dynamic locale key"
		}

		r := locale.ValueExpr.Range()
		edits = append(edits,
			sourceEdit{start: r.Start.Byte, end: r.Start.Byte, text: "jsonencode("},
			sourceEdit{start: r.End.Byte, end: r.End.Byte, text: heredocEnds[r.End.Byte] + ")"},
		)
	}

	if compact {
		edits = append(edits, sourceEdit{start: locales.Range().End.Byte, end: end})
	} else {
		// A trailing argument comma is valid in a call, but not in a group.
		afterObject := locales.Range().End
		suffix := src[afterObject.Byte:call.CloseParenRange.Start.Byte]

		tokens, _ := hclsyntax.LexExpression(suffix, call.Range().Filename, afterObject)
		for _, token := range tokens {
			if token.Type == hclsyntax.TokenComma {
				edits = append(edits, sourceEdit{start: token.Range.Start.Byte, end: token.Range.End.Byte})
			}
		}
	}

	return edits, ""
}

func staticObjectKey(expr hclsyntax.Expression) (string, bool) {
	value, diags := expr.Value(nil)
	if diags.HasErrors() || !value.IsKnown() || value.IsNull() || value.Type() != cty.String {
		return "", false
	}

	return value.AsString(), true
}
