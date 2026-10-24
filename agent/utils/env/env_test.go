package env

import (
	"testing"

	"github.com/joho/godotenv"
)

func TestMarshalKeepsValuesExactly(t *testing.T) {
	values := []string{"3306", "0", "-1", "007", "+123456", "-0123", "+0", "-0", "a'b", `a"b`, `a\b`}
	for _, v := range values {
		envMap := map[string]string{"PANEL_DB_ROOT_PASSWORD": v}
		for name, marshal := range map[string]func(map[string]string) (string, error){
			"Marshal": Marshal,
			"MarshalWithOrder": func(m map[string]string) (string, error) {
				return MarshalWithOrder(m, []string{"PANEL_DB_ROOT_PASSWORD"})
			},
		} {
			content, err := marshal(envMap)
			if err != nil {
				t.Fatalf("%s(%q) returned error: %v", name, v, err)
			}
			parsed, err := godotenv.Unmarshal(content)
			if err != nil {
				t.Fatalf("%s(%q) produced unparsable content %q: %v", name, v, content, err)
			}
			if got := parsed["PANEL_DB_ROOT_PASSWORD"]; got != v {
				t.Errorf("%s(%q) wrote %q, which reads back as %q", name, v, content, got)
			}
		}
	}
}

func TestMarshalWritesCanonicalIntegersUnquoted(t *testing.T) {
	content, err := Marshal(map[string]string{"PANEL_APP_PORT_HTTP": "8080"})
	if err != nil {
		t.Fatal(err)
	}
	if content != "PANEL_APP_PORT_HTTP=8080" {
		t.Fatalf("Marshal() = %q, want %q", content, "PANEL_APP_PORT_HTTP=8080")
	}
}
