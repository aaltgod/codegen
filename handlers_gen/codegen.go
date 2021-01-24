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
	Min string
	Max string
	Enum string
	Enums string
	Dflt string
}

var (
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
	paramStringTpl = template.Must(template.New("paramStringTpl").Parse(`
	output.{{.ParamName}} = r.FormValue("{{.Param}}")`))

	paramIntTpl = template.Must(template.New("paramIntTpl").Parse(`
	output.{{.ParamName}}, _ = strconv.Atoi(r.FormValue("{{.Param}}"))`))

	checkRequestMethod = template.Must(template.New("checkRequestMethod").Parse(`
func checkRequestMethod(availableMethod string, w http.ResponseWriter, r *http.Request) {
	if availableMethod == r.Method || availableMethod == "" {
		return
	}

	http.Error(w, "bad method", 406)
}`))

	checkForRequestParamTpl = template.Must(template.New("checkForRequestParamTpl").Parse(`
	if output.{{.ParamName}} == "" {
		http.Error(w, "{{.Param}} must me not empty", 400)
		return
	}`))

	checkForMinimumLenTpl = template.Must(template.New("checkForMinimumLenTpl").Parse(`
	if len(output.{{.ParamName}}) < {{.Min}} {
		http.Error(w, "{{.Param}} len must be >= {{.Min}}", 400)
		return
	}`))

	checkForMinimumNumberTpl = template.Must(template.New("checkForMinimumNumberTpl").Parse(`
	if output.{{.ParamName}} < {{.Min}} {
		http.Error(w, "{{.Param}} must be >= {{.Min}}", 400)
		return
	}`))

	checkForMaximumLenTpl = template.Must(template.New("checkForMaximumLenTpl").Parse(`
	if len(output.{{.ParamName}}) > {{.Max}} {
		http.Error(w, "{{.Param}} len must be <= {{.Min}}", 400)
		return
	}`))

	checkForMaximumNumberTpl = template.Must(template.New("checkForMaximumNumberTpl").Parse(`
	if output.{{.ParamName}} > {{.Max}} {
		http.Error(w, "{{.Param}} must be <= {{.Max}}", 400)
		return
	}`))

	enumSwitchTpl = template.Must(template.New("enumSwitchTpl").Parse(`
	if output.{{.ParamName}} != "" {
		switch output.{{.ParamName}} {`))

	enumCaseTpl = template.Must(template.New("enumCaseTpl").Parse(`
		case "{{.Enum}}":
			break`))

	enumDefaultTpl = template.Must(template.New("enumDefaultTpl").Parse(`
		default:
			http.Error(w, "{{.Param}} must be one of {{.Enums}}", 400)
			return
		}
	} else {
		output.{{.ParamName}} = "{{.Dflt}}"
	}`))
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

func getRequestMethod(methodName *ast.FuncDecl) string {
	comments := methodName.Doc.Text()
	matched, _ := regexp.MatchString(`"method": "[A-Z]*"`, comments)
	if !matched {
		return ""
	}

	s := regexp.MustCompile(`"method": "[A-Z]*"`).FindString(comments)
	method := regexp.MustCompile(`"`).Split(s, -1)[3]
	log.Println("METHHOD:", method)

	return method
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
	fmt.Fprintln(out, `import "strconv"`)
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
		fmt.Fprintln(out, `		http.Error(w, "", 404)`)
		fmt.Fprintln(out, `		return`)
		fmt.Fprintln(out, "	}\n}")

		for _, methodName := range methodAst {
			wrapperTpl.Execute(out, tpl{
				WrapperName: methodName.Name.Name+"Wrapper",
				StructName: structName,
			})

			log.Println("COMMENT:", methodName.Doc.Text())

			requestMethod := getRequestMethod(methodName)
			fmt.Fprintln(out, `	checkRequestMethod("`+requestMethod+`", w, r)`)
			fmt.Fprintln(out)


			for _, param := range methodName.Type.Params.List{
				paramName := fmt.Sprintf("%s", param.Type)

				for name, fields := range findStructs {
					if name != paramName {
						log.Println(name, "!=", paramName)
						continue
					}

					fmt.Fprintln(out, "	output := &" + paramName + "{}")

					for _, field := range fields {
						fieldName := field.Names[0].Name
						log.Println("FIELD:", fieldName)

						tag := reflect.StructTag(field.Tag.Value[1:len(field.Tag.Value)-1]).Get("apivalidator")
						log.Println("TAG:", tag)

						lowCaseFieldName := strings.ToLower(fieldName)
						matched, _ := regexp.MatchString("paramname=", tag)
						if matched {
							s := regexp.MustCompile(`paramname=[^,]*`).FindString(tag)
							lowCaseFieldName = regexp.MustCompile("=").Split(s, -1)[1]
						}

						fieldType := field.Type.(*ast.Ident).Name

						switch fieldType {
						case "string":
							paramStringTpl.Execute(out, tpl{
								ParamName: fieldName,
								Param:     lowCaseFieldName,
							})
						case "int":
							paramIntTpl.Execute(out, tpl{
								ParamName: fieldName,
								Param:     lowCaseFieldName,
							})

						}


						matched, _ = regexp.MatchString("required=", tag)
						if matched {
							checkForRequestParamTpl.Execute(out, tpl{
								ParamName: fieldName,
								Param:     lowCaseFieldName,
							})
						}

						matched, _ = regexp.MatchString("min=", tag)
						if matched {
							s := regexp.MustCompile(`min=[^,]*`).FindString(tag)
							min := regexp.MustCompile("=").Split(s, -1)[1]

							switch fieldType {
							case "string":
								checkForMinimumLenTpl.Execute(out, tpl{
									ParamName: fieldName,
									Param:     lowCaseFieldName,
									Min:       min,
								})
							case "int":
								checkForMinimumNumberTpl.Execute(out, tpl{
								ParamName: fieldName,
								Param:     lowCaseFieldName,
								Min:       min,
								})

							}
						}

						matched, _ = regexp.MatchString("max=", tag)
						if matched {
							s := regexp.MustCompile(`max=[^,]*`).FindString(tag)
							max := regexp.MustCompile("=").Split(s, -1)[1]

							switch fieldType {
							case "string":
								checkForMaximumLenTpl.Execute(out, tpl{
									ParamName: fieldName,
									Param:     lowCaseFieldName,
									Max:       max,
								})
							case "int":
								checkForMaximumNumberTpl.Execute(out, tpl{
									ParamName: fieldName,
									Param:     lowCaseFieldName,
									Max:       max,
								})

							}
						}

						matched, _ = regexp.MatchString("enum=", tag)
						if matched {

							enumSwitchTpl.Execute(out, tpl{
								ParamName: fieldName,
							})

							s := regexp.MustCompile(`enum=[^,]*`).FindString(tag)
							unparsedEnums := regexp.MustCompile("=").Split(s, -1)[1]
							log.Println("unparsedEnum:", unparsedEnums)


							enums := regexp.MustCompile("[|]").Split(unparsedEnums, -1)
							for _, enum := range enums {
								enumCaseTpl.Execute(out, tpl{
									Enum: enum,
								})
							}

							log.Println("ENUM:", enums)

							joinedEnums := "[" + strings.Join(enums, ", ") + "]"

							s = regexp.MustCompile(`default=[^,]*`).FindString(tag)
							dflt := regexp.MustCompile("=").Split(s, -1)[1]

							log.Println("DFLT:", dflt)

							enumDefaultTpl.Execute(out, tpl{
								ParamName: fieldName,
								Param: lowCaseFieldName,
								Enums: joinedEnums,
								Dflt: dflt,
							})
						}
					}
				}
			}
			fmt.Fprintln(out, "\n}")
		}

	}
	checkRequestMethod.Execute(out, tpl{})

}
