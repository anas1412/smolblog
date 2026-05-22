package main

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Date is a calendar date that serializes as YYYY-MM-DD in YAML frontmatter.
type Date struct {
	time.Time
}

func (d Date) MarshalYAML() (interface{}, error) {
	return d.Format("2006-01-02"), nil
}

func (d *Date) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date %q, expected YYYY-MM-DD: %w", s, err)
	}
	d.Time = t
	return nil
}

// parseFrontmatter reads a markdown document with standard YAML frontmatter.
//
//	---
//	title: My Article
//	date: 2026-05-22
//	slug: my-article
//	---
//	Markdown body…
func parseFrontmatter(raw []byte, dest interface{}) (body string, err error) {
	content := string(raw)

	if !strings.HasPrefix(content, "---\n") {
		return "", fmt.Errorf("missing opening frontmatter delimiter (---)")
	}

	rest := content[4:] // skip "---\n"
	endIdx := strings.Index(rest, "\n---\n")
	if endIdx == -1 {
		return "", fmt.Errorf("missing closing frontmatter delimiter (---)")
	}

	yamlBlock := rest[:endIdx]
	body = rest[endIdx+5:] // skip past "\n---\n"

	if err := yaml.Unmarshal([]byte(yamlBlock), dest); err != nil {
		return "", fmt.Errorf("yaml parse error: %w", err)
	}

	return strings.TrimSpace(body), nil
}

// renderFrontmatter serializes a YAML-marshalable struct and a body back to
// a markdown file with YAML frontmatter.
func renderFrontmatter(fm interface{}, body string) ([]byte, error) {
	yamlBlock, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("yaml marshal error: %w", err)
	}

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.WriteString(string(yamlBlock))
	buf.WriteString("---\n")
	buf.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		buf.WriteByte('\n')
	}
	return []byte(buf.String()), nil
}
