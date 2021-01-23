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
	StructName string
	SwitchName string
	WrapperName string
}


var (
	profileWrapper = "profileWrapper"
	createWrapper = "createWrapper"
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

	srvHTTPTpl = template.Must(template.New("srvTpl").Parse(`
// {{.StructName}}
func (srv *{{.StructName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
`))

	switchTpl = template.Must(template.New("switchTpl").Parse(`
	// {{.SwitchName}}
	switch r.URL.Path {
	case "/user/profile":
		srv.profileWrapper(w, r)
	case "/user/create":
		srv.createWrapper(w, r)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}`))

	wrapperTpl = template.Must(template.New("wrapperTpl").Parse(`
// {{.WrapperName}}
func (srv *{{.StructName}}) {{.WrapperName}}(w http.ResponseWriter, r *http.Request) {
`))
)

func main() {

	fSet := token.NewFileSet()
	node, err := parser.ParseFile(fSet, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])
	findMethods := make(map[string][]*ast.FuncDecl)
	findStructs := make(map[string][]string)

	fmt.Fprintln(out, `package ` + node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `import "net/http"`)
	//fmt.Fprintln(out, `import "encoding/binary"`)
	//fmt.Fprintln(out, `import "bytes"`)
	fmt.Fprintln(out)

	for _, f := range node.Decls {
		d, ok := f.(*ast.FuncDecl)
		if ok {
			if !strings.HasPrefix(d.Doc.Text(), "apigen:api"){
				continue
			}
			log.Println("\t\tFUNC:", d.Name, d.Type.Params)

			findStruct := d.Recv.List
			for _, spec := range findStruct {
				findStructType := spec.Type
				switch fst := findStructType.(type) {
				case *ast.StarExpr:
					findStructName := fmt.Sprintf("%s", fst.X)
					findMethods[findStructName] = append(findMethods[findStructName], d)

					log.Println(findMethods)
				}
			}
		} else {
			g, ok := f.(*ast.GenDecl)
			if !ok {
				log.Println("Type %T is not *ast.GenDecl", g)
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

				structName := currType.Name.Name

			FIELDSLOOP:
				for _, field := range currStruct.Fields.List {
					if field.Tag != nil {
						tag := reflect.StructTag(field.Tag.Value[1:len(field.Tag.Value)-1])
						if tag.Get("cgen") == "-" {
							continue FIELDSLOOP
						} else if value, exists := tag.Lookup("apivalidator"); exists {
							log.Printf("\tTAG: %s\n", value)
							fieldName := field.Names[0].Name
							findStructs[structName] = append(findStructs[structName], fieldName)
						}
					}
				}
			}
		}
	}

	for structName, methodAst := range findMethods {
		log.Println(structName, methodAst)
		srvHTTPTpl.Execute(out, tpl{
			StructName: structName,
		})
		switchTpl.Execute(out, tpl{
			SwitchName: structName+"Switch",
		})
		fmt.Fprintln(out, "\n}")
		wrapperTpl.Execute(out, tpl{
			WrapperName: createWrapper,
			StructName: structName,
		})
		fmt.Fprintln(out, "}")
		wrapperTpl.Execute(out, tpl{
			WrapperName: profileWrapper,
			StructName: structName,
		})
		fmt.Fprintln(out, "}")

	}
}
