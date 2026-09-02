/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

func ParseOpenAPISchema(document []byte, path string) (*entity.PhysicalSchema, error) {
	var doc map[string]any
	if json.Unmarshal(document, &doc) != nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	paths, _ := doc["paths"].(map[string]any)
	pathItem, _ := paths[path].(map[string]any)
	get, _ := pathItem["get"].(map[string]any)
	responses, _ := get["responses"].(map[string]any)
	keys := make([]string, 0, len(responses))
	for key := range responses {
		if key == "200" || (len(key) == 3 && key[0] == '2') {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return nil, entity.ErrDataDiscoveryFailed
	}
	response, _ := responses[keys[0]].(map[string]any)
	content, _ := response["content"].(map[string]any)
	media, _ := content["application/json"].(map[string]any)
	schema, _ := media["schema"].(map[string]any)
	if schema == nil {
		return nil, entity.ErrDataDiscoveryFailed
	}
	schema, err := resolveOpenAPIRef(doc, schema, map[string]bool{})
	if err != nil {
		return nil, err
	}
	if schema["type"] == "array" {
		schema, _ = schema["items"].(map[string]any)
		schema, err = resolveOpenAPIRef(doc, schema, map[string]bool{})
		if err != nil {
			return nil, err
		}
	}
	out := &entity.PhysicalSchema{Name: "GET " + path, Fields: []entity.PhysicalField{}, Relationships: []entity.PhysicalRelationship{}}
	flattenOpenAPISchema(doc, schema, "", true, &out.Fields)
	if len(out.Fields) == 0 {
		return nil, entity.ErrDataDiscoveryFailed
	}
	return out, nil
}

func resolveOpenAPIRef(doc map[string]any, schema map[string]any, seen map[string]bool) (map[string]any, error) {
	ref, _ := schema["$ref"].(string)
	if ref == "" {
		return schema, nil
	}
	if seen[ref] || !strings.HasPrefix(ref, "#/") {
		return nil, entity.ErrDataDiscoveryFailed
	}
	seen[ref] = true
	var current any = doc
	for _, token := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, entity.ErrDataDiscoveryFailed
		}
		current = object[strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")]
	}
	resolved, ok := current.(map[string]any)
	if !ok {
		return nil, entity.ErrDataDiscoveryFailed
	}
	return resolved, nil
}

func flattenOpenAPISchema(doc map[string]any, schema map[string]any, prefix string, required bool, fields *[]entity.PhysicalField) {
	resolved, err := resolveOpenAPIRef(doc, schema, map[string]bool{})
	if err != nil {
		return
	}
	if resolved["type"] == "array" {
		item, _ := resolved["items"].(map[string]any)
		flattenOpenAPISchema(doc, item, prefix, required, fields)
		return
	}
	properties, _ := resolved["properties"].(map[string]any)
	requiredSet := map[string]bool{}
	if names, ok := resolved["required"].([]any); ok {
		for _, name := range names {
			if value, ok := name.(string); ok {
				requiredSet[value] = true
			}
		}
	}
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		child, _ := properties[name].(map[string]any)
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		resolvedChild, err := resolveOpenAPIRef(doc, child, map[string]bool{})
		if err != nil {
			continue
		}
		kind, _ := resolvedChild["type"].(string)
		if kind == "object" || resolvedChild["properties"] != nil || kind == "array" {
			flattenOpenAPISchema(doc, resolvedChild, path, required && requiredSet[name], fields)
			continue
		}
		description, _ := resolvedChild["description"].(string)
		*fields = append(*fields, entity.PhysicalField{
			Name: path, Path: path, DataType: openAPIDataType(resolvedChild), NativeType: kind,
			Nullable: !(required && requiredSet[name]), Description: description, Ordinal: len(*fields) + 1,
		})
	}
}

func openAPIDataType(schema map[string]any) string {
	kind, _ := schema["type"].(string)
	format, _ := schema["format"].(string)
	if format != "" {
		return format
	}
	switch kind {
	case "integer", "number", "boolean", "string":
		return kind
	default:
		return "unknown"
	}
}
