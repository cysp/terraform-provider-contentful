package main

import (
	"flag"
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"
)

func main() {
	flag.Parse()

	patterns := flag.Args()

	pkgs, pkgsErr := packages.Load(&packages.Config{Mode: packages.LoadSyntax}, patterns...)
	if pkgsErr != nil {
		fmt.Println("Error loading packages", pkgsErr)
		return
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			fmt.Println("Errors loading package", pkg.Errors)
			continue
		}

		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				fmt.Println(decl)

				switch decl := decl.(type) {
				case *ast.GenDecl:
					if decl.Doc != nil {
						for _, comment := range decl.Doc.List {
							fmt.Println("decl doc", comment.Text)
						}
					}

					for _, spec := range decl.Specs {
						switch spec := spec.(type) {
						case *ast.TypeSpec:
							if spec.Doc != nil {
								for _, comment := range spec.Doc.List {
									fmt.Println("spec doc", comment.Text)
								}
							}
							if spec.Comment != nil {
								fmt.Println("spec comment", spec.Comment.Text())
							}

							switch specType := spec.Type.(type) {
							case *ast.InterfaceType:
								fmt.Println("interface", spec.Name.Name)
							case *ast.StructType:
								fmt.Println("struct", spec.Name.Name)

								for _, field := range specType.Fields.List {
									name := "<unnamed>"
									if len(field.Names) > 0 {
										name = field.Names[0].Name
									} else {
										if field.Type != nil {
											switch fieldType := field.Type.(type) {
											case *ast.Ident:
												name = fieldType.Name
											case *ast.SelectorExpr:
												name = fieldType.Sel.Name
												// name = fieldType.X.(*ast.Ident).Name + "." + name
											default:
												panic(fmt.Sprintf("unknown field type: %+v", field.Type))
											}
										}
									}

									tag := ""
									if field.Tag != nil {
										tag = field.Tag.Value
									}

									fmt.Println("field", name, tag)

									if field.Doc != nil {
										for _, comment := range field.Doc.List {
											fmt.Println("field doc", comment.Text)
										}
									}

									if field.Comment != nil {
										fmt.Println("field comment", field.Comment.Text())
									}
								}
							}
						}
					}
				}
			}
		}

		// for _, name := range pkg.Types.Scope().Names() {
		// 	fmt.Println(name)

		// 	obj := pkg.Types.Scope().Lookup(name)
		// 	if obj == nil {
		// 		fmt.Println("object not found")
		// 		continue
		// 	}

		// 	obj.Pos()

		// 	structType, ok := obj.Type().Underlying().(*types.Struct)
		// 	if !ok {
		// 		fmt.Println("not a struct")
		// 		continue
		// 	}

		// 	structType.
		// }
	}
}
