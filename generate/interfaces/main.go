package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	files, err := getFiles(".")
	if err != nil {
		panic(err)
	}

	var enforcers []*enforcer
	interfaces := make(map[string][]*ast.Field)
	for i := range files {
		file, err := parser.ParseFile(token.NewFileSet(), files[i], nil, 0)
		if err != nil {
			panic(err)
		}

		enforcers = append(enforcers, getAllEnforces(file)...)
		getAllInterfaces(file, interfaces)
	}

	for k, v := range interfaces {
		methods := ""
		for i := range v {
			if len(v[i].Names) == 0 {
				continue // TODO: embedded interfaces
			}
			methods += ", " + v[i].Names[0].Name
		}
		fmt.Println(k, methods)
	}
}

func getAllInterfaces(file *ast.File, interfaces map[string][]*ast.Field) {
	for name, item := range file.Scope.Objects {
		if item.Kind != ast.Typ {
			continue
		}

		// interface type
		typeDecl := item.Decl.(*ast.TypeSpec)
		var iDecl *ast.InterfaceType
		var ok bool
		if iDecl, ok = typeDecl.Type.(*ast.InterfaceType); !ok {
			continue
		}

		interfaces[name] = iDecl.Methods.List
	}
}

type enforcer struct {
	iName string // interface
	sName string // struct
}

func getAllEnforces(file *ast.File) (enforcers []*enforcer) {
	for _, item := range file.Decls {
		var gdecl *ast.GenDecl
		var ok bool
		if gdecl, ok = item.(*ast.GenDecl); !ok {
			continue
		}

		if gdecl.Tok != token.VAR {
			continue
		}

		specs := item.(*ast.GenDecl).Specs
		for i := range specs {
			vs := specs[i].(*ast.ValueSpec)
			if len(vs.Names) == 0 || vs.Names[0].Name != "_" {
				continue
			}

			var cExpr *ast.CallExpr
			if cExpr, ok = vs.Values[0].(*ast.CallExpr); !ok {
				continue
			}

			var pExpr *ast.ParenExpr
			if pExpr, ok = cExpr.Fun.(*ast.ParenExpr); !ok {
				continue
			}

			var sExpr *ast.StarExpr
			if sExpr, ok = pExpr.X.(*ast.StarExpr); !ok {
				continue
			}

			var id *ast.Ident
			if id, ok = sExpr.X.(*ast.Ident); !ok {
				continue
			}

			var id2 *ast.Ident
			if id2, ok = vs.Type.(*ast.Ident); !ok {
				continue
			}

			enforcers = append(enforcers, &enforcer{
				iName: id2.Name,
				sName: id.Name,
			})
		}
	}

	return enforcers
}

func getFiles(path string) (files []string, err error) {
	var results []string
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		results = append(results, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i := range results {
		isGoFile := strings.HasSuffix(results[i], ".go")
		isInSubDir := strings.Contains(results[i], "/")
		isGenFile := strings.HasSuffix(results[i], "_gen.go")
		if results[i] == path || !isGoFile || isInSubDir || isGenFile {
			continue
		}

		files = append(files, results[i])
	}

	return files, nil
}
