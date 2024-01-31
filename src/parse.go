package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	pbparse "github.com/yoheimuta/go-protoparser"
)

func parseProto(path string) ParseResult {
	reader, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to open %s, err %v\n", path, err)
	}
	defer reader.Close()

	got, err := pbparse.Parse(
		reader,
		pbparse.WithDebug(false),
		pbparse.WithPermissive(false),
		pbparse.WithFilename(filepath.Base(path)),
	)
	if err != nil {
		log.Fatal(err)
	}

	schema, err := pbparse.UnorderedInterpret(got)
	if err != nil {
		log.Fatal(err)
	}

	out := ParseResult{}

	for _, option := range schema.ProtoBody.Options {
		if option.OptionName == "go_package" {
			out.Pkg = strings.ReplaceAll(option.Constant, "\"", "")
		}
	}

	for _, mes := range schema.ProtoBody.Messages {
		msgDef := MessageDef{
			Name:   mes.MessageName,
			Fields: []FieldDef{},
		}
		for _, field := range mes.MessageBody.Fields {
			slicePrefix := ""
			if field.IsRepeated {
				slicePrefix = "[]"
			}
			if field.FieldNumber == "1" {
				msgDef.KeyField = field.FieldName
			}
			msgDef.Fields = append(msgDef.Fields, FieldDef{
				Name: field.FieldName,
				Type: slicePrefix + protoToGolangType(field.Type),
			})
		}
		out.Messages = append(out.Messages, msgDef)
	}
	return out
}

func protoToGolangType(original string) string {
	switch original {
	case "double":
		return "float64"
	case "float":
		return "float64"
	case "sint32":
		return "float64"
	case "sint64":
		return "float64"
	case "fixed32":
		return "float64"
	case "fixed64":
		return "float64"
	case "sfixed32":
		return "float64"
	case "sfixed64":
		return "float64"
	case "bytes":
		return "byte[]"
	}
	return original
}

type ParseResult struct {
	Pkg      string
	Messages []MessageDef
}

type MessageDef struct {
	Name     string
	KeyField string
	Fields   []FieldDef
}

type FieldDef struct {
	Name string
	Type string
}
