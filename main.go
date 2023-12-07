package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/goccy/go-yaml"
)

// TODO: topological sort for structs
// TODO: doxygen compatible docs for structs, enums, functions, function args
// TODO: strict validation of the yml, checking for references
// TODO: ability to generate header using two input yml files, impl-specific yml and standard yml

const (
	docsEnabled = true
)

var (
	inFlag  = flag.String("i", "", "yaml")
	outFlag = flag.String("o", "", "header")
)

type Yml struct {
	Copyright string     `yaml:"copyright"`
	Basetypes []string   `yaml:"basetypes"`
	Global    Global     `yaml:"global"`
	Enums     []Enum     `yaml:"enums"`
	Callbacks []Function `yaml:"callbacks"`
	Structs   []Struct   `yaml:"structs"`
	Objects   []Object   `yaml:"objects"`
}

type Global struct {
	Constants []GlobalConstants
}

type GlobalConstants struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
	Doc   string `yaml:"doc"`
}

type Enum struct {
	Name    string        `yaml:"name"`
	Doc     string        `yaml:"doc"`
	Bitmask bool          `yaml:"bitmask"`
	Entries []EnumEntries `yaml:"entries"`
}

type EnumEntries struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
	Doc   string `yaml:"doc"`
}

type Function struct {
	Name    string          `yaml:"name"`
	Doc     string          `yaml:"doc"`
	Returns FunctionReturns `yaml:"returns"`
	Args    []FunctionArg   `yaml:"args"`
}

type FunctionReturns struct {
	Doc     string      `yaml:"doc"`
	Type    string      `yaml:"type"`
	Pointer PointerType `yaml:"pointer"`
}

type FunctionArg struct {
	Name     string      `yaml:"name"`
	Doc      string      `yaml:"doc"`
	Type     string      `yaml:"type"`
	Pointer  PointerType `yaml:"pointer"`
	Optional bool        `yaml:"optional"`
}

type Struct struct {
	Name    string         `yaml:"name"`
	Type    string         `yaml:"type"`
	Doc     string         `yaml:"doc"`
	Members []StructMember `yaml:"members"`
}

type StructMember struct {
	Name     string      `yaml:"name"`
	Type     string      `yaml:"type"`
	Pointer  PointerType `yaml:"pointer"`
	Optional bool        `yaml:"optional"`
	Doc      string      `yaml:"doc"`
}

type PointerType string

const (
	PointerTypeMutable   PointerType = "mutable"
	PointerTypeImmutable PointerType = "immutable"
)

type Object struct {
	Name    string     `yaml:"name"`
	Doc     string     `yaml:"doc"`
	Methods []Function `yaml:"methods"`
}

func main() {
	flag.Parse()
	inPath := *inFlag
	outPath := *outFlag
	if inPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	srcData, err := os.ReadFile(inPath)
	if err != nil {
		panic(err)
	}

	var yml Yml
	if err := yaml.Unmarshal(srcData, &yml); err != nil {
		panic(err)
	}

	var out io.Writer
	if outPath == "" {
		out = os.Stdout
	} else {
		f, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		out = f
	}
	outBuffered := bufio.NewWriter(out)
	defer func() {
		if err := outBuffered.Flush(); err != nil {
			panic(err)
		}
	}()

	writeHeader(outBuffered, &yml)
}

func CValue(s string) string {
	switch s {
	case "usize_max":
		return "SIZE_MAX"
	case "uint32_max":
		return "0xffffffffUL"
	case "uint64_max":
		return "0xffffffffffffffffULL"
	default:
		return s
	}
}

func CType(typ string, pointerType PointerType) string {
	appendModifiers := func(s string, pointerType PointerType) string {
		var sb strings.Builder
		sb.WriteString(s)
		switch pointerType {
		case PointerTypeImmutable:
			sb.WriteString(" const *")
		case PointerTypeMutable:
			sb.WriteString(" *")
		}
		return sb.String()
	}
	switch typ {
	case "bool":
		return appendModifiers("WGPUBool", pointerType)
	case "string":
		return appendModifiers("char", PointerTypeImmutable)
	case "uint16":
		return appendModifiers("uint16_t", pointerType)
	case "uint32":
		return appendModifiers("uint32_t", pointerType)
	case "uint64":
		return appendModifiers("uint64_t", pointerType)
	case "usize":
		return appendModifiers("size_t", pointerType)
	case "int16":
		return appendModifiers("int16_t", pointerType)
	case "int32":
		return appendModifiers("int32_t", pointerType)
	case "float32":
		return appendModifiers("float", pointerType)
	case "float64":
		return appendModifiers("double", pointerType)
	case "c_void":
		return appendModifiers("void", pointerType)
	}

	switch {
	case strings.HasPrefix(typ, "enum."):
		ctype := "WGPU" + PascalCase(strings.TrimPrefix(typ, "enum."))
		return appendModifiers(ctype, pointerType)
	case strings.HasPrefix(typ, "struct."):
		ctype := "WGPU" + PascalCase(strings.TrimPrefix(typ, "struct."))
		return appendModifiers(ctype, pointerType)
	case strings.HasPrefix(typ, "callback."):
		ctype := "WGPU" + PascalCase(strings.TrimPrefix(typ, "callback."))
		return appendModifiers(ctype, pointerType)
	case strings.HasPrefix(typ, "object."):
		ctype := "WGPU" + PascalCase(strings.TrimPrefix(typ, "object."))
		return appendModifiers(ctype, pointerType)
	default:
		panic("TODO: " + typ)
	}
}

// MultilineComment converts a multiline string to a C-style multiline comment
// starting with '/**', ending with '*/' and each line prefixed with ' * '.
//
//	                 /**
//	Hello     =>      * Hello
//	World     =>      * World
//	                  */
func MultilineComment(in string, indent int) string {
	const space = ' '
	var out strings.Builder
	for i := 0; i < indent; i++ {
		out.WriteRune(space)
	}
	out.WriteString("/**\n")
	sc := bufio.NewScanner(strings.NewReader(strings.TrimSpace(in)))
	for sc.Scan() {
		line := sc.Text()
		for i := 0; i < indent; i++ {
			out.WriteRune(space)
		}
		out.WriteString(" * ")
		out.WriteString(line)
		out.WriteString("\n")
	}
	if err := sc.Err(); err != nil {
		panic(err)
	}
	for i := 0; i < indent; i++ {
		out.WriteRune(space)
	}
	out.WriteString(" */")
	return out.String()
}

// ConstantCase converts a string from snake case to constant case.
//
//	"whole_map_size" => "WHOLE_MAP_SIZE"
//	"whole_map_SIZE" => "WHOLE_MAP_SIZE"
func ConstantCase(v string) string {
	return strings.ToUpper(v)
}

// PascalCase converts a string from snake case to pascal case.
//
//	"whole_map_size" => "WholeMapSize"
//	"whole_map_SIZE" => "WholeMapSIZE"
func PascalCase(s string) string {
	var out strings.Builder
	out.Grow(len(s))
	capitalise := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if capitalise {
			out.WriteRune(unicode.ToUpper(rune(c)))
			capitalise = false
		} else {
			if c == '_' {
				capitalise = true
			} else {
				out.WriteRune(rune(c))
			}
		}
	}
	return out.String()
}

// CamelCase converts a string from snake case to camel case.
//
//	"whole_map_size" => "wholeMapSize"
//	"whole_map_SIZE" => "wholeMapSIZE"
func CamelCase(s string) string {
	var out strings.Builder
	out.Grow(len(s))
	capitalize := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if capitalize {
			out.WriteRune(unicode.ToUpper(rune(c)))
			capitalize = false
		} else {
			if c == '_' {
				capitalize = true
			} else {
				out.WriteRune(rune(c))
			}
		}
	}
	return out.String()
}
