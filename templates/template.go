package templates

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stelligent/mu/common"
	"gopkg.in/yaml.v2"
	"io"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// NewTemplate will create a temp file with the template for a CFN stack
//go:generate go-bindata -pkg $GOPACKAGE -o assets.go assets/
func NewTemplate(assetName string, data interface{}, cfnUpdates interface{}) (io.Reader, error) {
	asset, err := Asset(fmt.Sprintf("assets/%s", assetName))
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(assetName).Parse(string(asset[:]))
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	bufWriter := bufio.NewWriter(buf)

	err = tmpl.Execute(bufWriter, data)
	if err != nil {
		return nil, err
	}

	bufWriter.Flush()

	if cfnUpdates != nil {
		templateMap := make(map[interface{}]interface{})
		cleanYaml := fixupYaml(buf.Bytes())
		err = yaml.Unmarshal(cleanYaml, templateMap)
		if err != nil {
			return nil, newYamlError(err, cleanYaml)
		}
		common.MapApply(templateMap, cfnUpdates)
		yamlBytes, err := yaml.Marshal(templateMap)
		if err != nil {
			return nil, err
		}

		buf = bytes.NewBuffer(yamlBytes)
	}

	return buf, nil
}

func fixupYaml(yamlBytes []byte) []byte {
	scanner := bufio.NewScanner(bytes.NewReader(yamlBytes))

	buf := new(bytes.Buffer)
	bufWriter := bufio.NewWriter(buf)
	tagRegexp := regexp.MustCompile("^(\\s*)(.+?)!(\\S+)\\s+(.*?)$")
	fnTags := []string{"Ref", "If", "Not", "And", "Or", "Equals", "GetAtt", "Select", "Condition", "ImportValue", "GetAZs", "Base64", "FindInMap", "Sub", "Join", "Split"}

	extraIndentUntil := 0
	for scanner.Scan() {
		line := scanner.Text()
		matches := tagRegexp.FindStringSubmatch(line)
		extraIndent := ""
		if extraIndentUntil > 0 {
			if len(strings.TrimSpace(line)) == 0 {
				// no op
			} else if strings.HasPrefix(line, strings.Repeat(" ", extraIndentUntil+1)) {
				extraIndent = "  "
			} else {
				extraIndentUntil = 0
			}
		}
		if len(matches) > 0 {
			indent := matches[1]
			pre := matches[2]
			tag := matches[3]
			post := matches[4]

			for _, fn := range fnTags {
				if tag == fn {
					var tagWithPrefix string
					if tag == "Ref" || tag == "Condition" {
						tagWithPrefix = quoteString(tag)
					} else {
						tagWithPrefix = quoteString(fmt.Sprintf("Fn::%s", tag))
					}
					if post == "|" {
						line = fmt.Sprintf("%s%s\n%s  %s: %s", indent, pre, indent, tagWithPrefix, post)
						//add extra indent until we are back to indent is back to current level
						extraIndentUntil = len(indent)
					} else {
						line = fmt.Sprintf("%s%s {%s: %s}", indent, pre, tagWithPrefix, quoteString(post))
					}
				}
			}
		}
		bufWriter.WriteString(fmt.Sprintf("%s%s\n", extraIndent, line))

	}
	bufWriter.Flush()
	return buf.Bytes()
}

func quoteString(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") || strings.HasPrefix(s, "{") {
		return s
	}

	if !strings.HasPrefix(s, "\"") {
		s = fmt.Sprintf("\"%s", s)
	}
	if !strings.HasSuffix(s, "\"") {
		s = fmt.Sprintf("%s\"", s)
	}
	return s
}

func newYamlError(err error, yamlBytes []byte) error {
	errRegexp := regexp.MustCompile("line (\\d+)")
	matches := errRegexp.FindStringSubmatch(err.Error())
	lineNumber := -1
	if matches != nil {
		lineNumber, _ = strconv.Atoi(matches[1])
	}

	buf := new(bytes.Buffer)
	bufWriter := bufio.NewWriter(buf)
	scanner := bufio.NewScanner(bytes.NewReader(yamlBytes))
	num := 1
	for scanner.Scan() {
		line := scanner.Text()
		if lineNumber == -1 || lineNumber == num {
			bufWriter.WriteString(fmt.Sprintf("%d:\t%s\n", num, line))
		}
		num = num + 1
	}
	bufWriter.Flush()
	return errors.Wrap(err, buf.String())
}
