package gemini

import (
	"strings"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/relayconvert/kitutil"
)

var geminiOpenAPISchemaAllowedFields = map[string]struct{}{
	"anyOf":            {},
	"default":          {},
	"description":      {},
	"enum":             {},
	"example":          {},
	"format":           {},
	"items":            {},
	"maxItems":         {},
	"maxLength":        {},
	"maxProperties":    {},
	"maximum":          {},
	"minItems":         {},
	"minLength":        {},
	"minProperties":    {},
	"minimum":          {},
	"nullable":         {},
	"pattern":          {},
	"properties":       {},
	"propertyOrdering": {},
	"required":         {},
	"title":            {},
	"type":             {},
}

const geminiFunctionSchemaMaxDepth = 64

func CleanFunctionParameters(params interface{}) interface{} {
	return cleanGeminiFunctionParametersWithDepth(params, 0)
}

// CleanClaudeCompatibleFunctionParameters reduces a Gemini function schema to
// the conservative JSON Schema subset accepted by Gemini-to-Claude gateways.
func CleanClaudeCompatibleFunctionParameters(params interface{}) interface{} {
	cleaned := CleanFunctionParameters(params)
	result := cleanClaudeCompatibleSchema(cleaned, 0)
	if _, hasType := result["type"]; !hasType {
		result["type"] = "object"
	}
	return result
}

func cleanClaudeCompatibleSchema(schema interface{}, depth int) map[string]interface{} {
	if depth >= geminiFunctionSchemaMaxDepth {
		return map[string]interface{}{}
	}

	source, ok := schema.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}

	if alternatives, ok := source["anyOf"].([]interface{}); ok {
		for _, alternative := range alternatives {
			candidate := cleanClaudeCompatibleSchema(alternative, depth+1)
			if _, hasType := candidate["type"]; !hasType {
				continue
			}
			if description, ok := source["description"].(string); ok && description != "" {
				candidate["description"] = description
			}
			return candidate
		}
	}

	result := make(map[string]interface{})
	if description, ok := source["description"].(string); ok && description != "" {
		result["description"] = description
	}
	if schemaType, ok := source["type"].(string); ok {
		normalizedType := strings.ToLower(schemaType)
		switch normalizedType {
		case "object", "array", "string", "integer", "number", "boolean":
			result["type"] = normalizedType
		}
	}
	if enum, ok := source["enum"].([]interface{}); ok && len(enum) > 0 {
		result["enum"] = enum
	}

	if properties, ok := source["properties"].(map[string]interface{}); ok {
		cleanedProperties := make(map[string]interface{}, len(properties))
		for name, property := range properties {
			cleanedProperties[name] = cleanClaudeCompatibleSchema(property, depth+1)
		}
		result["properties"] = cleanedProperties
		result["type"] = "object"
	}
	if required, ok := source["required"].([]interface{}); ok {
		cleanedRequired := make([]string, 0, len(required))
		for _, item := range required {
			if name, ok := item.(string); ok && name != "" {
				cleanedRequired = append(cleanedRequired, name)
			}
		}
		if len(cleanedRequired) > 0 {
			result["required"] = cleanedRequired
		}
	}
	if items, exists := source["items"]; exists {
		result["items"] = cleanClaudeCompatibleSchema(items, depth+1)
		if _, hasType := result["type"]; !hasType {
			result["type"] = "array"
		}
	}

	return result
}

// CleanTools sanitizes the OpenAPI Schema fields embedded in native Gemini
// function declarations while preserving all other current and future tool
// types verbatim.
func CleanTools(tools []byte) ([]byte, error) {
	if len(tools) == 0 {
		return tools, nil
	}

	var value interface{}
	if err := kitutil.Unmarshal(tools, &value); err != nil {
		return nil, err
	}

	switch typed := value.(type) {
	case []interface{}:
		for _, tool := range typed {
			cleanGeminiTool(tool)
		}
	case map[string]interface{}:
		cleanGeminiTool(typed)
	}

	return kitutil.Marshal(value)
}

func cleanGeminiTool(tool interface{}) {
	toolMap, ok := tool.(map[string]interface{})
	if !ok {
		return
	}

	for _, key := range []string{"functionDeclarations", "function_declarations"} {
		declarations, ok := toolMap[key].([]interface{})
		if !ok {
			continue
		}
		for _, declaration := range declarations {
			declarationMap, ok := declaration.(map[string]interface{})
			if !ok {
				continue
			}
			for _, schemaKey := range []string{"parameters", "response"} {
				if schema, exists := declarationMap[schemaKey]; exists {
					declarationMap[schemaKey] = CleanFunctionParameters(schema)
				}
			}
		}
	}
}

func cleanGeminiFunctionParametersWithDepth(params interface{}, depth int) interface{} {
	if params == nil {
		return nil
	}

	if depth >= geminiFunctionSchemaMaxDepth {
		return cleanGeminiFunctionParametersShallow(params)
	}

	switch v := params.(type) {
	case map[string]interface{}:
		cleanedMap := make(map[string]interface{}, len(v))
		for key, val := range v {
			if _, ok := geminiOpenAPISchemaAllowedFields[key]; ok {
				cleanedMap[key] = val
			}
		}
		convertGeminiSchemaConst(v, cleanedMap)

		normalizeGeminiSchemaTypeAndNullable(cleanedMap)

		if props, ok := cleanedMap["properties"].(map[string]interface{}); ok && props != nil {
			cleanedProps := make(map[string]interface{})
			for propName, propValue := range props {
				cleanedProps[propName] = cleanGeminiFunctionParametersWithDepth(propValue, depth+1)
			}
			cleanedMap["properties"] = cleanedProps
		}

		if items, ok := cleanedMap["items"].(map[string]interface{}); ok && items != nil {
			cleanedMap["items"] = cleanGeminiFunctionParametersWithDepth(items, depth+1)
		}
		if itemsArray, ok := cleanedMap["items"].([]interface{}); ok && len(itemsArray) > 0 {
			cleanedMap["items"] = cleanGeminiFunctionParametersWithDepth(itemsArray[0], depth+1)
		}

		if nested, ok := cleanedMap["anyOf"].([]interface{}); ok && nested != nil {
			cleanedNested := make([]interface{}, len(nested))
			for i, item := range nested {
				cleanedNested[i] = cleanGeminiFunctionParametersWithDepth(item, depth+1)
			}
			cleanedMap["anyOf"] = cleanedNested
		}

		return cleanedMap
	case []interface{}:
		cleanedArray := make([]interface{}, len(v))
		for i, item := range v {
			cleanedArray[i] = cleanGeminiFunctionParametersWithDepth(item, depth+1)
		}
		return cleanedArray
	default:
		return params
	}
}

func cleanGeminiFunctionParametersShallow(params interface{}) interface{} {
	switch v := params.(type) {
	case map[string]interface{}:
		cleanedMap := make(map[string]interface{}, len(v))
		for key, val := range v {
			if _, ok := geminiOpenAPISchemaAllowedFields[key]; ok {
				cleanedMap[key] = val
			}
		}
		convertGeminiSchemaConst(v, cleanedMap)
		normalizeGeminiSchemaTypeAndNullable(cleanedMap)
		delete(cleanedMap, "properties")
		delete(cleanedMap, "items")
		delete(cleanedMap, "anyOf")
		return cleanedMap
	case []interface{}:
		return []interface{}{}
	default:
		return params
	}
}

func convertGeminiSchemaConst(source, target map[string]interface{}) {
	if _, hasEnum := target["enum"]; !hasEnum {
		if constValue, hasConst := source["const"]; hasConst {
			if stringValue, ok := constValue.(string); ok {
				target["enum"] = []interface{}{stringValue}
			}
		}
	}
}

func normalizeGeminiSchemaTypeAndNullable(schema map[string]interface{}) {
	rawType, ok := schema["type"]
	if !ok || rawType == nil {
		return
	}

	normalize := func(t string) (string, bool) {
		switch strings.ToLower(strings.TrimSpace(t)) {
		case "object":
			return "OBJECT", false
		case "array":
			return "ARRAY", false
		case "string":
			return "STRING", false
		case "integer":
			return "INTEGER", false
		case "number":
			return "NUMBER", false
		case "boolean":
			return "BOOLEAN", false
		case "null":
			return "", true
		default:
			return t, false
		}
	}

	switch typed := rawType.(type) {
	case string:
		normalized, isNull := normalize(typed)
		if isNull {
			schema["nullable"] = true
			delete(schema, "type")
			return
		}
		schema["type"] = normalized
	case []interface{}:
		nullable := false
		var chosen string
		for _, item := range typed {
			if value, ok := item.(string); ok {
				normalized, isNull := normalize(value)
				if isNull {
					nullable = true
					continue
				}
				if chosen == "" {
					chosen = normalized
				}
			}
		}
		if nullable {
			schema["nullable"] = true
		}
		if chosen != "" {
			schema["type"] = chosen
		} else {
			delete(schema, "type")
		}
	}
}

func RemoveAdditionalProperties(schema interface{}, depth int) interface{} {
	if depth >= 5 {
		return schema
	}

	value, ok := schema.(map[string]interface{})
	if !ok || len(value) == 0 {
		return schema
	}
	delete(value, "title")
	delete(value, "$schema")
	if typeVal, exists := value["type"]; !exists || (typeVal != "object" && typeVal != "array") {
		return schema
	}
	switch value["type"] {
	case "object":
		delete(value, "additionalProperties")
		if properties, ok := value["properties"].(map[string]interface{}); ok {
			for key, nested := range properties {
				properties[key] = RemoveAdditionalProperties(nested, depth+1)
			}
		}
		for _, field := range []string{"allOf", "anyOf", "oneOf"} {
			if nested, ok := value[field].([]interface{}); ok {
				for i, item := range nested {
					nested[i] = RemoveAdditionalProperties(item, depth+1)
				}
			}
		}
	case "array":
		if items, ok := value["items"].(map[string]interface{}); ok {
			value["items"] = RemoveAdditionalProperties(items, depth+1)
		}
	}

	return value
}

func OpenAIToolChoiceToConfig(toolChoice any) *dto.ToolConfig {
	if toolChoice == nil {
		return nil
	}

	if toolChoiceStr, ok := toolChoice.(string); ok {
		config := &dto.ToolConfig{
			FunctionCallingConfig: &dto.FunctionCallingConfig{},
		}
		switch toolChoiceStr {
		case "auto":
			config.FunctionCallingConfig.Mode = "AUTO"
		case "none":
			config.FunctionCallingConfig.Mode = "NONE"
		case "required":
			config.FunctionCallingConfig.Mode = "ANY"
		default:
			config.FunctionCallingConfig.Mode = "AUTO"
		}
		return config
	}

	if toolChoiceMap, ok := toolChoice.(map[string]interface{}); ok {
		if toolChoiceMap["type"] == "function" {
			config := &dto.ToolConfig{
				FunctionCallingConfig: &dto.FunctionCallingConfig{
					Mode: "ANY",
				},
			}
			if function, ok := toolChoiceMap["function"].(map[string]interface{}); ok {
				if name, ok := function["name"].(string); ok && name != "" {
					config.FunctionCallingConfig.AllowedFunctionNames = []string{name}
				}
			}
			return config
		}
		return nil
	}

	return nil
}
