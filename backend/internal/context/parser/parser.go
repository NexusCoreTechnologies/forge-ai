package parser

import (
    "encoding/json"
    "gopkg.in/yaml.v3"
)

// ParseYAML returns a minimal string representation of YAML content.
func ParseYAML(b []byte) (string, error) {
    var v any
    if err := yaml.Unmarshal(b, &v); err != nil {
        return "", err
    }
    out, _ := json.Marshal(v)
    return string(out), nil
}

// ParseJSON pretty prints JSON content.
func ParseJSON(b []byte) (string, error) {
    var v any
    if err := json.Unmarshal(b, &v); err != nil {
        return "", err
    }
    out, _ := json.MarshalIndent(v, "", "  ")
    return string(out), nil
}

// ParseText returns the text as-is.
func ParseText(b []byte) (string, error) {
    return string(b), nil
}
