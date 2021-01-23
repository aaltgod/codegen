package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

type tpl struct {
	FieldName string
	StructName string
	SwitchName string
	WrapperName string
	Path string
	MethodName string
	Param string
	ParamName string
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
func (srv *{{.StructName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {`))

	switchTpl = template.Must(template.New("switchTpl").Parse(`
	// {{.SwitchName}}
	switch r.URL.Path {`))

	caseTpl = template.Must(template.New("caseTpl").Parse(`
	case "{{.Path}}":
		srv.{{.MethodName}}Wrapper(w, r)`))

	wrapperTpl = template.Must(template.New("wrapperTpl").Parse(`
func (srv *{{.StructName}}) {{.WrapperName}}(w http.ResponseWriter, r *http.Request) {
`))
	paramTpl = template.Must(template.New("paramTpl").Parse(`
	{{.ParamName}} := r.URL.Query().Get("Param")`))
)

func getMethodsAndStructs(node *ast.File, findMethods map[string][]*ast.FuncDecl, findStructs map[string][]*ast.Field) {
	for _, f := range node.Decls {
		d, ok := f.(*ast.FuncDecl)
		if ok {
			if !strings.HasPrefix(d.Doc.Text(), "apigen:api"){
				continue
			}

			findStruct := d.Recv.List
			for _, spec := range findStruct {
				findStructType := spec.Type
				switch fst := findStructType.(type) {
				case *ast.StarExpr:
					findStructName := fmt.Sprintf("%s", fst.X)
					findMethods[findStructName] = append(findMethods[findStructName], d)
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
						}
						if _, exists := tag.Lookup("apivalidator"); exists {
							findStructs[structName] = append(findStructs[structName], field)
							continue FIELDSLOOP
						}
						if _, exists := tag.Lookup("json"); exists {
							findStructs[structName] = append(findStructs[structName], field)
						}
					}
				}
			}
		}
	}
}

func getURLFromComments(methodName *ast.FuncDecl) string {

	var url string

	for _, path := range methodName.Doc.List {
		s := regexp.MustCompile(`"/[^"]*"`).FindString(path.Text)
		url = regexp.MustCompile(`"`).Split(s, -1)[1]
	}

	return url
}


func main() {

	fSet := token.NewFileSet()
	node, err := parser.ParseFile(fSet, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])


	fmt.Fprintln(out, `package ` + node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out)

	findMethods := make(map[string][]*ast.FuncDecl)
	findStructs := make(map[string][]*ast.Field)

	getMethodsAndStructs(node, findMethods, findStructs)

	for structName, methodAst := range findMethods {
		srvHTTPTpl.Execute(out, tpl{
			StructName: structName,
		})
		switchTpl.Execute(out, tpl{
			SwitchName: structName+"Switch",
		})

		for _, methodName := range methodAst {
		FINDSTRUCTNAME:
			for _, spec := range methodName.Recv.List{
				switch typeOfStruct := spec.Type.(type) {
				case *ast.StarExpr:
					nameOfStruct := fmt.Sprintf("%s", typeOfStruct.X)
					if nameOfStruct != structName {
						log.Println(nameOfStruct, "!=", structName)
						continue FINDSTRUCTNAME
					}

					break FINDSTRUCTNAME
				}
			}

			caseTpl.Execute(out, tpl{
				Path: getURLFromComments(methodName),
				MethodName: methodName.Name.Name,
			})
		}

		fmt.Fprintln(out, "\n	default:")
		fmt.Fprintln(out, `		http.Error(w, "", http.StatusBadRequest)`)
		fmt.Fprintln(out, "	}\n}")

		for _, methodName := range methodAst {
			wrapperTpl.Execute(out, tpl{
				WrapperName: methodName.Name.Name+"Wrapper",
				StructName: structName,
			})

			for _, res := range methodName.Type.Results.List{
				switch resType := res.Type.(type) {
				case *ast.StarExpr:
					resName := fmt.Sprintf("%s", resType.X)

					for name, fields := range findStructs {
						if name != resName {
							log.Println(name, "!=", resName)
							continue
						}

						for _, field := range fields {
							log.Println("FIELD:", field.Names[0])
						}

					}
				}
			}



			log.Println(len(findStructs))

			fmt.Fprintln(out, "}")
		}
	}
}
