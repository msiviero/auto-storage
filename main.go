package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := parseArgs()

	files, err := os.ReadDir(args.directory)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		isProto := strings.HasSuffix(file.Name(), ".proto")
		if isProto {
			parseResult := ParseProto(filepath.Join(args.directory, file.Name()))
			chunks := strings.Split(parseResult.Pkg, "/")
			pkg := chunks[len(chunks)-1]
			codeGen := NewCodeGen(pkg)

			for _, message := range parseResult.Messages {
				codeGen.GenCtorFunction(message)
				codeGen.GenStoreIterface(message)
				codeGen.GenStoreStruct(message)
				codeGen.GenerateGetFunction(message)
				codeGen.GenerateSetFunction(message)
				codeGen.GenerateDelFunction(message)
				codeGen.GenerateListFunction(message)
				codeGen.GenerateIterateFunction(message)
			}
			err := codeGen.file.Save(fmt.Sprintf("%s/%s", parseResult.Pkg, strings.ToLower(strings.ReplaceAll(file.Name(), ".proto", "_storage.g.go"))))

			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func parseArgs() ParsedArgs {
	inputDir := flag.String("d", "", "dir")
	flag.Parse()

	if (len(*inputDir)) < 1 {
		log.Fatal("No input directory provided")
	}

	file, err := os.Open(*inputDir)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if !fileInfo.IsDir() {
		log.Fatalf("%s is not a directory", *inputDir)
	}
	return ParsedArgs{directory: *inputDir}
}

type ParsedArgs struct {
	directory string
}
