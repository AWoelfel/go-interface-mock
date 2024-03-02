package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"text/template"
)

const DefaultAssertMockName = "Assert%sCall"

var (
	noSelfImport   = flag.Bool("noSelfImport", false, "skip import of packge containing <typeName>; might be useful if mock resides in same package")
	typeName       = flag.String("type", "", "target type name; must be set; type must be an interface")
	output         = flag.String("output", "", "output file name; default srcdir/<type>_mock.go")
	assertMockName = flag.String("asserName", DefaultAssertMockName, "mock name format; i.e. for `Assert%sCall` will create `AssertFunctionCall` for a method named 'Function'; defaults to `Assert%sCall`")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	_, _ = fmt.Fprintf(os.Stderr, "go-interface-mock is a lightweight code generator for interface mock.\n")
	_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	_, _ = fmt.Fprintf(os.Stderr, "\tgo-interface-mock [flags] -type T [directory]\n")
	_, _ = fmt.Fprintf(os.Stderr, "\tgo-interface-mock [flags] -type T files... # Must be a single package\n")
	_, _ = fmt.Fprintf(os.Stderr, "For more information, see:\n")
	_, _ = fmt.Fprintf(os.Stderr, "\thttp://godoc.org/github.com/AWoelfel/go-interface-mock\n")
	_, _ = fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

// parsePackage analyzes the single package constructed from the patterns and tags.
// parsePackage exits if there is an error.
func parsePackage(packagePath string) *packages.Package {
	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
		// TODO: Need to think about constants in test files. Maybe write type_string_test.go
		// in a separate pass? For later.
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, packagePath)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	return pkgs[0]
}

func templateToBuffer(writer io.Writer, template *template.Template, data any) error {

	err := template.Execute(writer, data)
	if err != nil {
		return fmt.Errorf("unable to execute template %s (%w)", template.Name(), err)
	}

	return nil
}

func must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}

type variable struct {
	*varPackageDep
	qualifier   types.Qualifier
	Name        string
	Enabled     bool
	OrderNumber int
	typ         types.Type
}

func returnCast(rc variable, s string) string {
	knownCasts := map[string]func(string) string{
		"error": func(s string) string { return fmt.Sprintf("utils.ToError(%s)", s) },
	}

	result, found := knownCasts[rc.typ.String()]

	if !found {

		switch vTyped := rc.typ.(type) {
		case *types.Pointer:
			result = func(s string) string {
				return fmt.Sprintf("utils.ToPointer[%s](%s)", types.TypeString(vTyped.Elem(), rc.qualifier), s)
			}

		default:
			result = func(s string) string {
				return fmt.Sprintf("%s.(%s)", s, types.TypeString(rc.typ, rc.qualifier))
			}
		}

	}

	return result(s)
}

type varPackageDep struct {
	codePackage *types.Package
	Pkg         string
	Alias       string
}

func resolveDependencyPackage(deps *[]*varPackageDep, p *types.Package) *varPackageDep {

	if p != nil {
		if i := slices.IndexFunc(*deps, func(e *varPackageDep) bool { return e.Pkg == p.Path() }); i < 0 {
			*deps = append(*deps, &varPackageDep{Pkg: p.Path(), Alias: p.Name(), codePackage: p})
		} else {
			return (*deps)[i]
		}
	}

	return nil

}

type signature struct {
	InVars    []variable
	OutVars   []variable
	qualifier types.Qualifier
	Signature *types.Signature
}

type signatureSet struct {
	ImplName string
	MockName string
	Impl     signature
	Mock     signature
}

func isEnabled(vType types.Type) bool {
	disabledTypes := []string{"context.Context"}
	return !slices.Contains(disabledTypes, vType.String())
}

func unwrapType(homePackage *types.Package, vType types.Type, deps *[]*varPackageDep, varName string) variable {
	switch vTyped := vType.(type) {

	//Enabled

	case *types.Named:
		return variable{
			varPackageDep: resolveDependencyPackage(deps, vTyped.Obj().Pkg()),
			Name:          varName,
			typ:           vTyped,
			Enabled:       isEnabled(vType),
			qualifier:     codeGenQualifier(homePackage, deps),
		}

	case *types.Basic:
		return variable{
			varPackageDep: nil,
			Name:          varName,
			typ:           vTyped,
			Enabled:       isEnabled(vType),
			qualifier:     codeGenQualifier(homePackage, deps),
		}

	case *types.Interface:
		return variable{
			varPackageDep: nil,
			Name:          varName,
			typ:           vTyped,
			Enabled:       isEnabled(vType),
			qualifier:     codeGenQualifier(homePackage, deps),
		}

	case *types.Pointer:
		res := unwrapType(homePackage, vTyped.Elem(), deps, varName)
		res.typ = vType
		return res

	case *types.Array:
		res := unwrapType(homePackage, vTyped.Elem(), deps, varName)
		res.typ = vType
		return res

	case *types.Slice:
		res := unwrapType(homePackage, vTyped.Elem(), deps, varName)
		res.typ = vType
		return res
	}

	panic(errors.New("unknown type"))
}

func codeGenQualifier(homePackage *types.Package, deps *[]*varPackageDep) types.Qualifier {

	relativeQualifier := types.RelativeTo(homePackage)

	return func(p *types.Package) string {

		if res := relativeQualifier(p); len(res) > 0 {
			i := slices.IndexFunc(*deps, func(dep *varPackageDep) bool { return dep.Pkg == p.Path() })
			return (*deps)[i].Alias
		}

		return ""
	}
}

func buildSignature(sig signature) string {
	buf := bytes.Buffer{}
	types.WriteSignature(&buf, sig.Signature, sig.qualifier)
	return buf.String()
}

func enabledInVars(sig signature) bool {
	return slices.ContainsFunc(sig.InVars, func(v variable) bool { return v.Enabled })
}

func mockSignature(homePackage *types.Package, originalSig *types.Signature, deps *[]*varPackageDep) (result signature) {

	argumentOrderNumber := 0
	unknownArgumentIdx := 0

	var allVars []*types.Var

	processParams := func(tup *types.Tuple, target *[]variable, unknownFormat string, format string) {
		for i := 0; i < tup.Len(); i++ {
			v := tup.At(i)
			vType := v.Type()

			varName := v.Name()
			if len(varName) == 0 {
				unknownArgumentIdx++
				varName = fmt.Sprintf(unknownFormat, unknownArgumentIdx)
			}

			unwrappedType := unwrapType(homePackage, vType, deps, varName)
			if unwrappedType.Enabled {
				unwrappedType.OrderNumber = argumentOrderNumber
				argumentOrderNumber++
				allVars = append(allVars, types.NewParam(v.Pos(), v.Pkg(), varName, v.Type()))
			}
			*target = append(*target, unwrappedType)
		}
	}

	processParams(originalSig.Params(), &result.InVars, "_%03d", "%s")
	processParams(originalSig.Results(), &result.OutVars, "out%03d", "out%s")

	result.Signature = types.NewSignatureType(nil, nil, nil, types.NewTuple(allVars...), nil, originalSig.Variadic())
	result.qualifier = codeGenQualifier(homePackage, deps)
	return
}

func implSignature(homePackage *types.Package, originalSig *types.Signature, deps *[]*varPackageDep) (result signature) {

	argumentOrderNumber := 0
	unknownArgumentIdx := 0

	var paramVars []*types.Var
	var resultVars []*types.Var

	repackVar := func(v *types.Var, target *[]variable, varName string) *types.Var {
		vType := v.Type()

		unwrappedType := unwrapType(homePackage, vType, deps, varName)
		if unwrappedType.Enabled {
			unwrappedType.OrderNumber = argumentOrderNumber
			argumentOrderNumber++
		}
		*target = append(*target, unwrappedType)
		return types.NewParam(v.Pos(), v.Pkg(), varName, v.Type())
	}

	for i := 0; i < originalSig.Params().Len(); i++ {
		v := originalSig.Params().At(i)

		varName := v.Name()
		if len(varName) == 0 {
			unknownArgumentIdx++
			varName = fmt.Sprintf("_%03d", unknownArgumentIdx)
		}
		paramVars = append(paramVars, repackVar(v, &result.InVars, varName))
	}

	for i := 0; i < originalSig.Results().Len(); i++ {
		v := originalSig.Results().At(i)
		resultVars = append(resultVars, repackVar(v, &result.OutVars, ""))
	}

	result.Signature = types.NewSignatureType(nil, nil, nil, types.NewTuple(paramVars...), types.NewTuple(resultVars...), originalSig.Variadic())
	result.qualifier = codeGenQualifier(homePackage, deps)
	return
}

func collectMethodVars(homePackage *types.Package, m *types.Func, deps *[]*varPackageDep) *signatureSet {

	if !m.Exported() {
		return nil
	}

	originalSig := m.Type().(*types.Signature)

	return &signatureSet{
		ImplName: m.Name(),
		MockName: fmt.Sprintf(*assertMockName, m.Name()),
		Impl:     implSignature(homePackage, originalSig, deps),
		Mock:     mockSignature(homePackage, originalSig, deps),
	}
}

const codeTemplate = `
package {{.package}}

import (
{{range .imports}} {{.Alias}}  "{{.Pkg}}"
{{end}}
)

func New{{.targetInterface}}Mock(t *testing.T) *{{.targetInterface}}Mock {
	result := {{.targetInterface}}Mock{t: t, MockedCalls: utils.NewMockedCalls()}
	t.Cleanup(func() { result.AssertNoCallsLeft(t) })
	return &result
}

type {{.targetInterface}}Mock struct {
	t *testing.T
	utils.MockedCalls
}

{{range .signatures}}

func (mockInstance *{{$.targetInterface}}Mock) {{.ImplName}}{{call $.buildSignature .Impl}}{
	{{if call $.enabledInVars .Impl}}idx{{else}}_{{end}}, objects := mockInstance.Next(mockInstance.t, "{{.ImplName}}")

	{{range .Impl.InVars}}
		{{if .Enabled}}
	assert.EqualValuesf(mockInstance.t, objects[{{.OrderNumber}}], {{.Name}}, "{{.Name}} miss match in call #%d", idx)
		{{end}}
	{{end}}
	return  {{range $i, $e := .Impl.OutVars}}{{if $i}}, {{end}}{{call $.returnCast . (printf "objects[%d]" .OrderNumber)}}{{end}}
}

func (mockInstance *{{$.targetInterface}}Mock) {{.MockName}}{{call $.buildSignature .Mock}} {
	mockInstance.AppendCall("{{.ImplName}}"{{range .Mock.InVars}}{{if .Enabled}}, {{.Name}}{{end}}{{end}}{{range .Mock.OutVars}}, {{.Name}}{{end}})
}

{{end}}
`

func main() {
	log.SetFlags(0)
	log.SetPrefix("go-interface-mock: ")
	flag.Usage = Usage
	flag.Parse()
	if len(*typeName) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	if len(args) > 1 {
		fmt.Printf("Only a single package is supported\n")
		os.Exit(2)
	}

	if len(*assertMockName) == 0 {
		*assertMockName = DefaultAssertMockName
	}

	if len(*output) == 0 {
		*output = filepath.ToSlash(filepath.Clean(fmt.Sprintf("%s/%s_mock.go", args[0], *typeName)))
	}

	pkg := parsePackage(args[0])

	scope := pkg.Types.Scope()

	target := scope.Lookup(*typeName)

	var dependencies []*varPackageDep
	var methodSignatures []signatureSet

	if target == nil {
		fmt.Printf("Interface %s not found in package\n", *typeName)
		os.Exit(3)
	}
	interfaceType, success := target.Type().Underlying().(*types.Interface)
	if !success {
		fmt.Printf("Object %s not an interface\n", *typeName)
		os.Exit(4)
	}

	for i := 0; i < interfaceType.NumMethods(); i++ {
		m := interfaceType.Method(i)
		resultSignatures := collectMethodVars(pkg.Types, m, &dependencies)

		methodSignatures = append(methodSignatures, *resultSignatures)
	}

	//add "testing" && "utils"
	resolveDependencyPackage(&dependencies, types.NewPackage("testing", "testing"))
	resolveDependencyPackage(&dependencies, types.NewPackage("github.com/stretchr/testify/assert", "assert"))
	resolveDependencyPackage(&dependencies, types.NewPackage("github.com/AWoelfel/go-interface-mock/utils", "utils"))

	if *noSelfImport {
		dependencies = slices.DeleteFunc(dependencies, func(dep *varPackageDep) bool { return dep.codePackage == pkg.Types })
	}

	data := map[string]any{}
	data["targetInterface"] = *typeName
	data["package"] = pkg.Types.Name()
	data["imports"] = dependencies
	data["signatures"] = methodSignatures
	data["returnCast"] = returnCast
	data["buildSignature"] = buildSignature
	data["enabledInVars"] = enabledInVars

	outBuffer := bytes.Buffer{}

	tmpl := must(template.New("tmpl").Parse(codeTemplate))
	err := tmpl.Execute(&outBuffer, data)
	if err != nil {
		panic(fmt.Errorf("unable to execute template %s (%w)", tmpl.Name(), err))
	}
	err = os.WriteFile(*output, outBuffer.Bytes(), os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("unable to write file %s (%w)", *output, err))
	}

}
