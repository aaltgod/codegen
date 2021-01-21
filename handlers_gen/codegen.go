package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

type tpl struct {
	FieldName string
}

var (
	intTpl = template.Must(template.New("intTpl").Parse(`
	// {{.FieldName}}
	var {{.FieldName}}Raw uint32
	binary.Read(reader, binary.LittleEndian, &{{.FieldName}}Raw)
	srv.{{.FieldName}} = int({{.FieldName}}Raw)
`))

	strTpl = template.Must(template.New("strTpl").Parse(`
	// {{.FieldName}}
	var {{.FieldName}}LenRaw uint32
	binary.Read(reader, binary.LittleEndian, &{{.FieldName}}LenRaw)
	{{.FieldName}}Raw := make([]byte, {{.FieldName}}LenRaw)
	binary.Read(reader, binary.LittleEndian, {{.FieldName}}Raw)
	srv.{{.FieldName}} = string({{.FieldName}}Raw)
`))
)

func main() {

	fmt.Println("I'm doing nothing")

	fSet := token.NewFileSet()
	node, err := parser.ParseFile(fSet, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package ` + node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out, `import "encoding/binary"`)
	fmt.Fprintln(out, `import "bytes"`)
	fmt.Fprintln(out)

	intTpl.Execute(out, tpl{"URL"})
	strTpl.Execute(out, tpl{FieldName:"Query"})

	for _, f := range node.Decls {
		d, ok := f.(*ast.FuncDecl)
		if !ok {
			log.Printf("SKIP %T is not *ast.FuncDel\n", d)
			continue
		}

		if !strings.HasPrefix(d.Doc.Text(), "apigen:api"){
			continue
		}
		log.Println("\t\tFUNC:", d.Name, d.Type.Params)

		rec := d.Recv.List
		for _, spec := range rec {
			fieldType := spec.Type
			log.Printf("%v", fieldType)
			switch f := fieldType.(type){
			case *ast.StarExpr:
				structName := f.X
				fmt.Fprintf(out, `func (srv *%v) ServeHTTP(w http.ResponseWriter, r *http.Request) {`, structName)
				fmt.Fprintln(out)
				fmt.Fprintln(out, ` reader := bytes.NewReader(r)`)
				log.Println("STRUCT NAME: ", structName)
			default:
				log.Println("unsupported structName")
				continue
			}


		}
		fmt.Fprintln(out,`}`)
		g, ok := f.(*ast.GenDecl)
		if !ok {
			log.Printf("SKIP %T is not *ast.GenDecl\n", f)
			continue
		}

		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)

			if !ok {
				log.Printf("SKIP %T is not *ast.TypeSpec\n", currType)
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				log.Print("SKIP %T is not *ast.StructType\n", currStruct)
				continue
			}


		FIELDSLOOP:
			for _, field := range currStruct.Fields.List {
				if field.Tag != nil {
					tag := reflect.StructTag(field.Tag.Value[1:len(field.Tag.Value)-1])
					if tag.Get("cgen") == "-" {
						continue FIELDSLOOP
					} else {
						fmt.Printf("%T\n", field.Tag.Value)
					}
				}

				fieldName := field.Names[0].Name
				fieldType := field.Type
				fmt.Printf("TYPE: %T\n", fieldType)

				switch ft := fieldType.(type) {
				case *ast.Ident:
					log.Println("IDENT", ft, fieldName)
					name := ft.Name
					log.Println("\tName: ", name)
					switch name {
					case "int":
						intTpl.Execute(out, tpl{FieldName: fieldName})
					case "string":
						strTpl.Execute(out, tpl{FieldName: fieldName})
					default:
						log.Println("unsupported", fieldType)
					}
				case *ast.StarExpr:
					log.Println("STAREXPR: ", ft, fieldName)
					star := ft.X
					log.Println("\tStar: ", star)
				case *ast.MapType:
					log.Println("MAPTYPE: ", ft, fieldName)
					key, value := ft.Key, ft.Value
					log.Println("\tMAP: ", key, value)
				default:
					log.Println("--------------------------", ft, fieldName)
				}
				//fmt.Printf("\t generating code for: %s.%s\n", currType.Name.Name, fieldName)
			}
		}
	}
}