package main

import (
	"fmt"
	"strings"
	"unicode"

	j "github.com/dave/jennifer/jen"
)

type CodeGen struct {
	file *j.File
}

func NewCodeGen(packageName string) CodeGen {
	f := j.NewFile(packageName)

	f.HeaderComment("Code generated by auto-storage. DO NOT EDIT.")

	f.ImportNames(map[string]string{
		"log":                              "",
		"fmt":                              "",
		"errors":                           "",
		"github.com/google/uuid":           "uuid",
		"github.com/cockroachdb/pebble":    "pebble",
		"google.golang.org/protobuf/proto": "proto",
		"github.com/gofiber/fiber/v2":      "fiber",
	})

	return CodeGen{file: f}
}

func (gen *CodeGen) GenCtorFunction(definition MessageDef) {
	gen.file.Func().
		Id(fmt.Sprintf("New%sStore", definition.Name)).
		Params(j.Id("path").String()).
		Op("*").Id(fmt.Sprintf("%sStore", definition.Name)).
		Block(
			j.Id("name").Op(":=").Qual("fmt", "Sprintf").Call(
				j.Lit("%s/%s"),
				j.Id("path"),
				j.Lit(strings.ToLower(definition.Name)),
			),
			j.List(j.Id("db"), j.Err()).Op(":=").Qual("github.com/cockroachdb/pebble", "Open").
				Call(
					j.Id("name"),
					j.Op("&").Qual("github.com/cockroachdb/pebble", "Options").Values(),
				),
			j.If(j.Id("err").Op("!=").Nil()).Block(
				j.Qual("log", "Fatal").Call(j.Id("err")),
			),
			j.Return().
				Op("&").Id(fmt.Sprintf("%sStore", definition.Name)).Values(j.Dict{j.Id("db"): j.Id("db")}),
		).Line()
}

func (gen *CodeGen) GenerateDataStruct(definition MessageDef) {
	typeFields := []j.Code{
		j.Id(capitalize("key")).Qual("github.com/google/uuid", "UUID").Tag(map[string]string{"json": "key"}),
	}
	for _, field := range definition.Fields {
		typeFields = append(typeFields,
			j.Id(capitalize(field.Name)).
				Id(field.Name).
				Tag(map[string]string{"json": strings.ToLower(field.Name)}),
		)
	}
	gen.file.Type().Id(definition.Name).Struct(typeFields...).Line()
}

func (gen *CodeGen) GenStoreStruct(definition MessageDef) {
	gen.file.Type().Id(fmt.Sprintf("%sStore", definition.Name)).Struct(
		j.Id("db").Id("*pebble.DB"),
	).Line()
}

func (gen *CodeGen) GenStoreIterface(definition MessageDef) {
	gen.file.Type().Id(fmt.Sprintf("%sStoreInterface", definition.Name)).Interface(
		j.Id("Get").Params(j.Op("*").Id(definition.Name)).Params(
			j.Op("*").Id(definition.Name),
			j.Error(),
		),
		j.Id("Set").Params(j.Op("*").Id(definition.Name)).Error(),
		j.Id("Delete").Params(j.Op("*").Id(definition.Name)).Error(),
		j.Id("List").Params().Params(j.Op("[]").Op("*").Id(definition.Name), j.Error()),
		j.Id("Iterate").Params(j.Func().Params(j.Op("*").Id(definition.Name))).Error(),
	).Line()
}

func (gen *CodeGen) GenerateGetFunction(definition MessageDef) {
	storeType := fmt.Sprintf("%sStore", definition.Name)
	gen.file.Func().
		Params(j.Id("store").Op("*").Id(storeType)).
		Id("Get").
		Params(j.Id("message").Op("*").Id(definition.Name)).
		Params(j.Op("*").Id(definition.Name), j.Error()).
		Block(
			j.Id("item").Op(":=").Op("&").Id(definition.Name).Values(),
			j.List(
				j.Id("bytes"),
				j.Id("closer"),
				j.Err(),
			).Op(":=").Id("store").Dot("db").Dot("Get").Call(j.Id("[]byte").Call(j.Id("message").Dot(capitalize(definition.KeyField)))),
			j.If(j.Qual("errors", "Is").Params(j.Err(), j.Qual("github.com/cockroachdb/pebble", "ErrNotFound"))).Block(j.Return(j.Nil(), j.Nil())),
			checkAndReturnNilAndError(),
			j.Err().Op("=").Qual("google.golang.org/protobuf/proto", "Unmarshal").Call(j.Id("bytes"), j.Id("item")),
			checkAndReturnNilAndError(),
			j.Err().Op("=").Id("closer").Dot("Close").Call(),
			checkAndReturnNilAndError(),
			j.Return(j.List(j.Id("item"), j.Nil())),
		).Line()
}

func (gen *CodeGen) GenerateSetFunction(definition MessageDef) {
	storeType := fmt.Sprintf("%sStore", definition.Name)
	gen.file.Func().
		Params(j.Id("store").Op("*").Id(storeType)).
		Id("Set").
		Params(j.Id("message").Op("*").Id(definition.Name)).
		Params(j.Error()).
		Block(
			j.List(
				j.Id("bytes"),
				j.Err(),
			).Op(":=").Qual("google.golang.org/protobuf/proto", "Marshal").
				Call(j.Id("message")),
			checkAndReturnError(),
			j.Err().Op("=").Id("store").Dot("db").Dot("Set").Call(
				j.Id("[]byte").Call(j.Id("message").Dot(capitalize(definition.KeyField))),
				j.Id("bytes"),
				j.Op("&").Qual("github.com/cockroachdb/pebble", "WriteOptions").Values(),
			),
			checkAndReturnError(),
			j.Id("store").Dot("db").Dot("Flush").Call(),
			j.Return(j.Nil()),
		).
		Line()
}

func (gen *CodeGen) GenerateDelFunction(definition MessageDef) {
	storeType := fmt.Sprintf("%sStore", definition.Name)
	gen.file.Func().
		Params(j.Id("store").Op("*").Id(storeType)).
		Id("Delete").
		Params(j.Id("message").Op("*").Id(definition.Name)).
		Params(j.Error()).
		Block(
			j.Err().Op(":=").Id("store").Dot("db").Dot("Delete").Call(
				j.Id("[]byte").Call(j.Id("message").Dot(capitalize(definition.KeyField))),
				j.Op("&").Qual("github.com/cockroachdb/pebble", "WriteOptions").Values(),
			),
			checkAndReturnError(),
			j.Id("store").Dot("db").Dot("Flush").Call(),
			j.Return(j.Nil()),
		).
		Line()
}

func (gen *CodeGen) GenerateListFunction(definition MessageDef) {
	storeType := fmt.Sprintf("%sStore", definition.Name)
	gen.file.Func().
		Params(j.Id("store").Op("*").Id(storeType)).
		Id("List").
		Params().
		Params(j.Id("[]").Op("*").Id(definition.Name), j.Error()).
		Block(
			j.Id("items").Op(":=").Id("[]").Op("*").Id(definition.Name).Values(),
			j.List(
				j.Id("iter"),
				j.Err(),
			).Op(":=").Id("store").Dot("db").Dot("NewIter").Call(
				j.Op("&").Qual("github.com/cockroachdb/pebble", "IterOptions").Values(),
			),
			checkAndReturnNilAndError(),
			j.For(
				j.Id("iter").Dot("First").Call(),
				j.Id("iter").Dot("Valid").Call(),
				j.Id("iter").Dot("Next").Call(),
			).Block(
				j.Id("item").Op(":=").Op("&").Id(definition.Name).Values(),
				j.Err().Op("=").Qual("google.golang.org/protobuf/proto", "Unmarshal").Call(
					j.Id("iter").Dot("Value").Call(),
					j.Id("item"),
				),
				j.If(j.Id("err").Op("!=").Nil()).Block(
					j.Break(),
				),
				j.Id("items").Op("=").Id("append").Call(j.Id("items"), j.Id("item")),
			),
			j.If(j.Id("err").Op("!=").Nil()).Block(
				j.Return(j.Nil(), j.Id("err")),
			),
			j.Return(j.Id("items"), j.Nil()),
		).
		Line()
}

func (gen *CodeGen) GenerateIterateFunction(definition MessageDef) {
	storeType := fmt.Sprintf("%sStore", definition.Name)
	gen.file.Func().
		Params(j.Id("store").Op("*").Id(storeType)).
		Id("Iterate").
		Params(j.Id("fn").Func().Params(j.Op("*").Id(definition.Name))).
		Params(j.Error()).
		Block(
			j.List(
				j.Id("iter"),
				j.Err(),
			).Op(":=").Id("store").Dot("db").Dot("NewIter").Call(
				j.Op("&").Qual("github.com/cockroachdb/pebble", "IterOptions").Values(),
			),
			checkAndReturnError(),
			j.For(
				j.Id("iter").Dot("First").Call(),
				j.Id("iter").Dot("Valid").Call(),
				j.Id("iter").Dot("Next").Call(),
			).Block(
				j.Id("item").Op(":=").Op("&").Id(definition.Name).Values(),
				j.Err().Op(":=").Qual("google.golang.org/protobuf/proto", "Unmarshal").Call(
					j.Id("iter").Dot("Value").Call(),
					j.Id("item"),
				),
				checkAndReturnError(),
				j.Id("fn").Call(j.Id("item")),
			),
			j.Return(j.Nil()),
		).
		Line()
}

func (gen *CodeGen) GenerateOneFunction(definition MessageDef) {
	storeType := fmt.Sprintf("%sStore", definition.Name)
	gen.file.Func().
		Params(j.Id("store").Op("*").Id(storeType)).
		Id("Iterate").
		Params(j.Id("fn").Func().Params(j.Op("*").Id(definition.Name))).
		Params(j.Error()).
		Block(
			j.Id("iter").Op(":=").Id("store").Dot("db").Dot("NewIter").Call(
				j.Op("&").Qual("github.com/cockroachdb/pebble", "IterOptions").Values(),
			),
			j.For(
				j.Id("iter").Dot("First").Call(),
				j.Id("iter").Dot("Valid").Call(),
				j.Id("iter").Dot("Next").Call(),
			).Block(
				j.Id("item").Op(":=").Op("&").Id(definition.Name).Values(),
				j.Err().Op(":=").Qual("google.golang.org/protobuf/proto", "Unmarshal").Call(
					j.Id("iter").Dot("Value").Call(),
					j.Id("item"),
				),
				checkAndReturnError(),
				j.Id("fn").Call(j.Id("item")),
			),
			j.Return(j.Nil()),
		).
		Line()
}

func (gen *CodeGen) PrintDebug() {
	fmt.Printf("%#v", gen.file)
}

func checkAndReturnError() *j.Statement {
	statement := j.If(j.Id("err").Op("!=").Nil()).Block(
		j.Return(j.List(j.Err())),
	)
	return statement
}

func checkAndReturnNilAndError() *j.Statement {
	statement := j.If(j.Id("err").Op("!=").Nil()).Block(
		j.Return(j.List(j.Nil(), j.Err())),
	)
	return statement
}

func capitalize(str string) string {
	runes := []rune(strings.ToLower(str))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
