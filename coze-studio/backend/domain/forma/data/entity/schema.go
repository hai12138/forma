/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

type PhysicalSchema struct {
	Name          string                 `json:"name"`
	Fields        []PhysicalField        `json:"fields"`
	Relationships []PhysicalRelationship `json:"relationships"`
}

type PhysicalField struct {
	Name        string         `json:"name"`
	DataType    string         `json:"data_type"`
	NativeType  string         `json:"native_type,omitempty"`
	Nullable    bool           `json:"nullable"`
	PrimaryKey  bool           `json:"primary_key"`
	Description string         `json:"description,omitempty"`
	Path        string         `json:"path,omitempty"`
	Ordinal     int            `json:"ordinal"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type PhysicalRelationship struct {
	Name             string   `json:"name"`
	FromFields       []string `json:"from_fields"`
	ToSchema         string   `json:"to_schema"`
	ToFields         []string `json:"to_fields"`
	RelationshipType string   `json:"relationship_type"`
}
