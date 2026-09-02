package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOpenAPISchemaReferencesNestedAndArrays(t *testing.T) {
	document := []byte(`{
	  "openapi":"3.0.3",
	  "paths":{
	    "/customers/{id}":{"get":{"responses":{"200":{"content":{"application/json":{"schema":{"$ref":"#/components/schemas/Customer"}}}}}}},
	    "/orders":{"get":{"responses":{"200":{"content":{"application/json":{"schema":{"type":"array","items":{"$ref":"#/components/schemas/Order"}}}}}}}}
	  },
	  "components":{"schemas":{
	    "Address":{"type":"object","required":["city"],"properties":{"city":{"type":"string","description":"City"}}},
	    "Customer":{"type":"object","required":["id","address"],"properties":{"id":{"type":"integer","format":"int64"},"address":{"$ref":"#/components/schemas/Address"}}},
	    "Order":{"type":"object","properties":{"id":{"type":"string"},"total":{"type":"number","format":"double"}}}
	  }}
	}`)
	customer, err := ParseOpenAPISchema(document, "/customers/{id}")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"id", "address.city"}, []string{customer.Fields[0].Name, customer.Fields[1].Name})
	for _, field := range customer.Fields {
		if field.Name == "address.city" {
			require.False(t, field.Nullable)
		}
	}
	orders, err := ParseOpenAPISchema(document, "/orders")
	require.NoError(t, err)
	require.Equal(t, []string{"id", "total"}, []string{orders.Fields[0].Name, orders.Fields[1].Name})
}
