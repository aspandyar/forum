package docs

import _ "embed"

// OpenAPISpec is the embedded OpenAPI 3 document (see openapi.yaml).
//
//go:embed openapi.yaml
var OpenAPISpec []byte
