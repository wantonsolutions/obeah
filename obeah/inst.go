package obeah

import (
	"bitbucket.org/bestchai/dinv/programslicer"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"log"
	"strings"
)

const CHARSTART = 913

var (
	debug     = false
	Directory = ""
	File      = ""
	Pipe      = ""
	logger    *log.Logger
)

func Insturment(options map[string]string, l *log.Logger) map[string]string {
	logger = l
	logger.SetPrefix("[Obeah Instrument] ")
	initalize(options)
	p, err := getProgramWrapper()
	if err != nil {
		panic(err)
	}

	instrumentedOutput := make(map[string]string)
	for pnum, pack := range p.Packages {
		for snum, _ := range pack.Sources {
			for _, cfg := range p.Packages[pnum].Sources[snum].Cfgs {
				fmt.Println("PRINTING CFGs")
				fmt.Println(cfg.Cfg.String(p.Fset, func(s ast.Stmt) string {
					return "(test)"
				}))
			}
			instSource := InstrumentSource(p.Fset, p.Packages[pnum].Sources[snum].Comments)
			p.Packages[pnum].Sources[snum].Text = instSource
			instrumentedOutput[p.Packages[pnum].Sources[snum].Filename] = instSource
		}
	}

	return instrumentedOutput
}

func InstrumentSource(fset *token.FileSet, file *ast.File) string {
	lines := ControlFlowLines(fset, file)
	buf := new(bytes.Buffer)
	printer.Fprint(buf, fset, file)
	split := strings.SplitAfter(buf.String(), "\n")
	mergedSource := make([]string, 0)
	id := 0
	for i := range split {
		mergedSource = append(mergedSource, split[i])
		if lines[i+1] {
			//mergedSource = append(mergedSource,fmt.Sprintf("obeah.Log(`%d`)\n",i+1))
			mergedSource = append(mergedSource, "obeah.Log(\""+string(id+CHARSTART)+"\")\n")
			id++
		}
	}
	instrumented := mergeSource(mergedSource)
	fmt.Println(instrumented)
	formatted, err := format.Source([]byte(instrumented))
	if err != nil {
		panic(err)
	}
	return string(formatted)
}

func ControlFlowLines(fset *token.FileSet, file *ast.File) map[int]bool {
	lines := make(map[int]bool, 0)
	ast.Inspect(file, func(n ast.Node) bool {
		switch c := n.(type) {
		case *ast.BlockStmt, *ast.CaseClause:
			switch c.(type) {
			case *ast.SelectStmt, *ast.SwitchStmt:
				break
			default:
				lines[fset.Position(c.Pos()).Line] = true
				break
			}
			break
		}
		return true
	})
	return lines
}

func mergeSource(source []string) string {
	var output string
	for _, line := range source {
		output = output + line
	}
	return output
}

func initalize(options map[string]string) {
	for setting := range options {
		switch setting {
		case "debug":
			debug = true
		case "directory":
			Directory = options[setting]
		case "file":
			File = options[setting]
		case "pipe":
			Pipe = options[setting]
		default:
			continue
		}
	}
}

func getProgramWrapper() (*programslicer.ProgramWrapper, error) {
	var (
		program *programslicer.ProgramWrapper
		err     error
	)
	if Directory != "" {
		program, err = programslicer.GetProgramWrapperDirectory(Directory)
		if err != nil {
			return program, err
		}
	} else if File != "" {
		program, err = programslicer.GetProgramWrapperFile(File)
		if err != nil {
			return program, err
		}
	} else if Pipe != "" {
		program, err = programslicer.GetWrapperFromString(Pipe)
		if err != nil {
			return program, err
		}
	}
	return program, nil
}
