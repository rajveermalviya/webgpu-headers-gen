package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"slices"
	"strconv"
	"strings"
)

func writeHeader(w io.Writer, yml *Yml) {
	writePreamble(w, yml)
	writeConstants(w, yml)
	writeTypeAliases(w, yml)
	writeForwardDecls(w, yml)
	writeEnums(w, yml)
	writeCallback(w, yml)
	writeStructs(w, yml)
	writeProcs(w, yml)
	writeFooter(w, yml)
}

func writePreamble(w io.Writer, yml *Yml) {
	fmt.Fprintln(w, MultilineComment(yml.Copyright, 0))
	fmt.Fprint(w, `
#ifndef WEBGPU_H_
#define WEBGPU_H_

#if defined(WGPU_SHARED_LIBRARY)
#    if defined(_WIN32)
#        if defined(WGPU_IMPLEMENTATION)
#            define WGPU_EXPORT __declspec(dllexport)
#        else
#            define WGPU_EXPORT __declspec(dllimport)
#        endif
#    else  // defined(_WIN32)
#        if defined(WGPU_IMPLEMENTATION)
#            define WGPU_EXPORT __attribute__((visibility("default")))
#        else
#            define WGPU_EXPORT
#        endif
#    endif  // defined(_WIN32)
#else       // defined(WGPU_SHARED_LIBRARY)
#    define WGPU_EXPORT
#endif  // defined(WGPU_SHARED_LIBRARY)

#if !defined(WGPU_OBJECT_ATTRIBUTE)
#define WGPU_OBJECT_ATTRIBUTE
#endif
#if !defined(WGPU_ENUM_ATTRIBUTE)
#define WGPU_ENUM_ATTRIBUTE
#endif
#if !defined(WGPU_STRUCTURE_ATTRIBUTE)
#define WGPU_STRUCTURE_ATTRIBUTE
#endif
#if !defined(WGPU_FUNCTION_ATTRIBUTE)
#define WGPU_FUNCTION_ATTRIBUTE
#endif
#if !defined(WGPU_NULLABLE)
#define WGPU_NULLABLE
#endif

#include <stdint.h>
#include <stddef.h>

`)
}

func writeForwardDecls(w io.Writer, yml *Yml) {
	objects := slices.Clone(yml.Objects)
	slices.SortStableFunc(objects, func(a, b Object) int {
		return strings.Compare(PascalCase(a.Name), PascalCase(b.Name))
	})
	for _, obj := range objects {
		name := PascalCase(obj.Name)
		fmt.Fprintf(w, "typedef struct WGPU%sImpl* WGPU%s WGPU_OBJECT_ATTRIBUTE;\n", name, name)
	}
	fmt.Fprintln(w)

	structs := slices.Clone(yml.Structs)
	slices.SortStableFunc(structs, func(a, b Struct) int {
		return strings.Compare(PascalCase(a.Name), PascalCase(b.Name))
	})
	for _, s := range structs {
		name := PascalCase(s.Name)
		fmt.Fprintf(w, "struct WGPU%s;\n", name)
	}
	fmt.Fprintln(w)
}

func writeConstants(w io.Writer, yml *Yml) {
	constants := yml.Global.Constants
	for _, constant := range constants {
		if docsEnabled {
			fmt.Fprintln(w, MultilineComment(constant.Doc, 0))
		}
		fmt.Fprintf(w, "#define WGPU_%s (%s)\n", ConstantCase(constant.Name), CValue(constant.Value))
	}
}

func writeTypeAliases(w io.Writer, yml *Yml) {
	fmt.Fprint(w, `
typedef uint32_t WGPUFlags;
typedef uint32_t WGPUBool;

`)
}

func writeEnums(w io.Writer, yml *Yml) {
	enums := slices.Clone(yml.Enums)
	slices.SortStableFunc(enums, func(a, b Enum) int {
		return strings.Compare(PascalCase(a.Name), PascalCase(b.Name))
	})
	for _, enum := range enums {
		if enum.Bitmask {
			continue
		}
		enumName := PascalCase(enum.Name)
		if docsEnabled {
			fmt.Fprintln(w, MultilineComment(enum.Doc, 0))
		}
		fmt.Fprintf(w, "typedef enum WGPU%s {\n", enumName)
		for entryIndex, entry := range enum.Entries {
			if docsEnabled {
				fmt.Fprintln(w, MultilineComment(entry.Doc, 4))
			}
			if entry.Value != "" {
				v, err := strconv.ParseUint(entry.Value, 10, 64)
				if err != nil {
					panic(err)
				}
				fmt.Fprintf(w, "    WGPU%s_%s = 0x%.8X,\n", enumName, PascalCase(entry.Name), v)
			} else {
				fmt.Fprintf(w, "    WGPU%s_%s = 0x%.8X,\n", enumName, PascalCase(entry.Name), entryIndex)
			}
		}
		fmt.Fprintf(w, "    WGPU%s_Force32 = 0x7FFFFFFF\n", enumName)
		fmt.Fprintf(w, "} WGPU%s WGPU_ENUM_ATTRIBUTE;\n\n", enumName)
	}

	for _, enum := range enums {
		if !enum.Bitmask {
			continue
		}
		enumName := PascalCase(enum.Name)
		if docsEnabled {
			fmt.Fprintln(w, MultilineComment(enum.Doc, 0))
		}
		fmt.Fprintf(w, "typedef enum WGPU%s {\n", enumName)
		for entryIndex, entry := range enum.Entries {
			if docsEnabled {
				fmt.Fprintln(w, MultilineComment(entry.Doc, 4))
			}
			if entry.Value != "" {
				v, err := strconv.ParseUint(entry.Value, 10, 64)
				if err != nil {
					panic(err)
				}
				fmt.Fprintf(w, "    WGPU%s_%s = 0x%.8X,\n", enumName, PascalCase(entry.Name), v)
			} else {
				fmt.Fprintf(w, "    WGPU%s_%s = 0x%.8X,\n", enumName, PascalCase(entry.Name), uint64(math.Pow(2, float64(entryIndex-1))))
			}
		}
		fmt.Fprintf(w, "    WGPU%s_Force32 = 0x7FFFFFFF\n", enumName)
		fmt.Fprintf(w, "} WGPU%s WGPU_ENUM_ATTRIBUTE;\n", enumName)
		fmt.Fprintf(w, "typedef WGPUFlags WGPU%sFlags WGPU_ENUM_ATTRIBUTE;\n\n", enumName)
	}
}

func writeCallback(w io.Writer, yml *Yml) {
	callbacks := slices.Clone(yml.Callbacks)
	slices.SortStableFunc(callbacks, func(a, b Function) int {
		return strings.Compare(PascalCase(a.Name), PascalCase(b.Name))
	})
	for _, cb := range callbacks {
		if docsEnabled {
			fmt.Fprintln(w, MultilineComment(cb.Doc, 0))
		}
		returnType := ""
		switch cb.Returns.Type {
		case "c_void":
			returnType = "void"
		default:
			panic(fmt.Sprintf("%#v", cb))
		}
		name := PascalCase(cb.Name)
		args := []string{}
		for _, arg := range cb.Args {
			v := ""
			if arg.Optional {
				v += "WGPU_NULLABLE "
			}
			v += CType(arg.Type, arg.Pointer)
			v += " " + CamelCase(arg.Name)
			args = append(args, v)
		}
		if len(args) == 0 {
			fmt.Fprintf(w, "typedef %s (*WGPU%s)(void) WGPU_FUNCTION_ATTRIBUTE;\n", returnType, name)
		} else {
			fmt.Fprintf(w, "typedef %s (*WGPU%s)(%s) WGPU_FUNCTION_ATTRIBUTE;\n", returnType, name, strings.Join(args, ", "))
		}
	}
}

func writeStructs(w io.Writer, yml *Yml) {
	fmt.Fprint(w, `
typedef struct WGPUChainedStruct {
    struct WGPUChainedStruct const * next;
    WGPUSType sType;
} WGPUChainedStruct WGPU_STRUCTURE_ATTRIBUTE;

typedef struct WGPUChainedStructOut {
    struct WGPUChainedStructOut * next;
    WGPUSType sType;
} WGPUChainedStructOut WGPU_STRUCTURE_ATTRIBUTE;

`)

	structs := slices.Clone(yml.Structs)
	slices.SortStableFunc(structs, func(a, b Struct) int {
		return strings.Compare(PascalCase(a.Name), PascalCase(b.Name))
	})

	var skippedStructs []Struct

mainloop:
	for _, s := range structs {
		for _, m := range s.Members {
			if strings.HasPrefix(m.Type, "struct.") {
				skippedStructs = append(skippedStructs, s)
				continue mainloop
			}
		}

		structName := PascalCase(s.Name)
		if docsEnabled {
			fmt.Fprintln(w, MultilineComment(s.Doc, 0))
		}
		fmt.Fprintf(w, "typedef struct WGPU%s {\n", structName)
		switch s.Type {
		case "standalone":
		case "base_in":
			fmt.Fprintln(w, "    WGPUChainedStruct const * nextInChain;")
		case "extension_in":
			fmt.Fprintln(w, "    WGPUChainedStruct chain;")
		case "base_out":
			fmt.Fprintln(w, "    WGPUChainedStructOut * nextInChain;")
		case "extension_out":
			fmt.Fprintln(w, "    WGPUChainedStructOut chain;")
		default:
			log.Panic("TODO")
		}
		for _, member := range s.Members {
			if docsEnabled {
				fmt.Fprintln(w, MultilineComment(member.Doc, 4))
			}
			if member.Optional {
				fmt.Fprintf(w, "    WGPU_NULLABLE %s %s;\n", CType(member.Type, member.Pointer), CamelCase(member.Name))
			} else {
				fmt.Fprintf(w, "    %s %s;\n", CType(member.Type, member.Pointer), CamelCase(member.Name))
			}
		}
		fmt.Fprintf(w, "} WGPU%s WGPU_STRUCTURE_ATTRIBUTE;\n\n", structName)
	}

	for _, s := range skippedStructs {
		structName := PascalCase(s.Name)
		if docsEnabled {
			fmt.Fprintln(w, MultilineComment(s.Doc, 0))
		}
		fmt.Fprintf(w, "typedef struct WGPU%s {\n", structName)
		switch s.Type {
		case "standalone":
		case "base_in":
			fmt.Fprintln(w, "    WGPUChainedStruct const * nextInChain;")
		case "extension_in":
			fmt.Fprintln(w, "    WGPUChainedStruct chain;")
		case "base_out":
			fmt.Fprintln(w, "    WGPUChainedStructOut * nextInChain;")
		case "extension_out":
			fmt.Fprintln(w, "    WGPUChainedStructOut chain;")
		default:
			log.Panic("TODO")
		}
		for _, member := range s.Members {
			if docsEnabled {
				fmt.Fprintln(w, MultilineComment(member.Doc, 4))
			}
			if member.Optional {
				fmt.Fprintf(w, "    WGPU_NULLABLE %s %s;\n", CType(member.Type, member.Pointer), CamelCase(member.Name))
			} else {
				fmt.Fprintf(w, "    %s %s;\n", CType(member.Type, member.Pointer), CamelCase(member.Name))
			}
		}
		fmt.Fprintf(w, "} WGPU%s WGPU_STRUCTURE_ATTRIBUTE;\n\n", structName)
	}
}

func writeProcs(w io.Writer, yml *Yml) {
	fmt.Fprintf(w, `
#ifdef __cplusplus
extern "C" {
#endif

#if !defined(WGPU_SKIP_PROCS)
`)

	objects := slices.Clone(yml.Objects)
	slices.SortStableFunc(objects, func(a, b Object) int {
		return strings.Compare(PascalCase(a.Name), PascalCase(b.Name))
	})

	for _, o := range objects {
		fmt.Fprintln(w)
		objectName := PascalCase(o.Name)
		fmt.Fprintf(w, "// Procs of %s\n", objectName)

		methods := slices.Clone(o.Methods)
		slices.SortStableFunc(methods, func(a, b Function) int {
			return strings.Compare(PascalCase(a.Name), PascalCase(b.Name))
		})
		methods = append(methods,
			Function{
				Name: "reference",
				Returns: FunctionReturns{
					Type: "c_void",
				},
			},
			Function{
				Name: "release",
				Returns: FunctionReturns{
					Type: "c_void",
				},
			},
		)

		for _, method := range methods {
			returnType := CType(method.Returns.Type, method.Returns.Pointer)
			name := PascalCase(method.Name)
			args := []string{
				fmt.Sprintf("WGPU%s %s", objectName, CamelCase(o.Name)),
			}

			for _, arg := range method.Args {
				v := ""
				if arg.Optional {
					v += "WGPU_NULLABLE "
				}
				v += CType(arg.Type, arg.Pointer)
				v += " " + CamelCase(arg.Name)
				args = append(args, v)
			}
			fmt.Fprintf(w, "typedef %s (*WGPUProc%s%s)(%s) WGPU_FUNCTION_ATTRIBUTE;\n", returnType, objectName, name, strings.Join(args, ", "))
		}
	}

	fmt.Fprintf(w, `
#endif  // !defined(WGPU_SKIP_PROCS)

#if !defined(WGPU_SKIP_DECLARATIONS)

`)

	for _, o := range objects {
		fmt.Fprintln(w)
		objectName := PascalCase(o.Name)
		fmt.Fprintf(w, "// Methods of %s\n", objectName)

		methods := slices.Clone(o.Methods)
		slices.SortStableFunc(methods, func(a, b Function) int {
			return strings.Compare(PascalCase(a.Name), PascalCase(b.Name))
		})
		methods = append(methods,
			Function{
				Name: "reference",
				Returns: FunctionReturns{
					Type: "c_void",
				},
			},
			Function{
				Name: "release",
				Returns: FunctionReturns{
					Type: "c_void",
				},
			},
		)

		for _, method := range methods {
			returnType := CType(method.Returns.Type, method.Returns.Pointer)
			name := PascalCase(method.Name)
			args := []string{
				fmt.Sprintf("WGPU%s %s", objectName, CamelCase(o.Name)),
			}

			for _, arg := range method.Args {
				v := ""
				if arg.Optional {
					v += "WGPU_NULLABLE "
				}
				v += CType(arg.Type, arg.Pointer)
				v += " " + CamelCase(arg.Name)
				args = append(args, v)
			}
			fmt.Fprintf(w, "WGPU_EXPORT %s wgpu%s%s(%s) WGPU_FUNCTION_ATTRIBUTE;\n", returnType, objectName, name, strings.Join(args, ", "))
		}
	}

	fmt.Fprintf(w, `
#endif  // !defined(WGPU_SKIP_DECLARATIONS)

#ifdef __cplusplus
} // extern "C"
#endif

`)
}

func writeFooter(w io.Writer, yml *Yml) {
	fmt.Fprintf(w, "#endif // WEBGPU_H_\n")
}
