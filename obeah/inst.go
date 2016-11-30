package obeah

import (
	"bitbucket.org/bestchai/dinv/programslicer"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
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

type Flow struct {
	Id        string
	Line      int
	Child     []*Flow
	Condition string
	Node      ast.Node
}

func NewFlow() *Flow {
	return &Flow{Line: -1, Child: make([]*Flow, 0), Condition: "", Node: nil}
}

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
	_, lines := ControlFlowLines(fset, file)
	buf := new(bytes.Buffer)
	printer.Fprint(buf, fset, file)
	split := strings.SplitAfter(buf.String(), "\n")
	mergedSource := make([]string, 0)
	id := 0
	for i := range split {
		mergedSource = append(mergedSource, split[i])
		if _, ok := lines[i+1]; ok {
			//mergedSource = append(mergedSource,fmt.Sprintf("obeah.Log(`%d`)\n",i+1))
			marker := fmt.Sprintf("%s-%d", lines[i+1].Id, lines[i+1].Line)
			mergedSource = append(mergedSource, "obeah.Log(\""+marker+"\")\n")
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

func ControlFlowLines(fset *token.FileSet, file *ast.File) (*Flow, map[int]Flow) {
	head := new(Flow)
	mapper := make(map[int]Flow)
	ast.Inspect(file, func(n ast.Node) bool {
		switch c := n.(type) {
		//function entrance
		case *ast.FuncDecl:
			f := NewFlow()
			f.Id = "TEST"
			f.Line = fset.Position(c.Body.Pos()).Line
			f.Node = n
			mapper[f.Line] = *f
			head = f
			//if statement //must come before
		case *ast.IfStmt:
			f := NewFlow()
			f.Id = "TEST"
			f.Line = fset.Position(c.Body.Pos()).Line
			f.Node = n
			f.Condition = nodeToString(c.Cond)
			parent, err := findParent(n, head, file, fset)
			if err != nil {
				panic(err)
			}
			parent.Child = append(parent.Child, f)
			fmt.Println(f.Condition)
			mapper[f.Line] = *f
			break
		case *ast.BlockStmt:
			//check for else

			if ok := isElse(n, file, fset); ok {
				print()
			}
			break
		}
		return true
	})
	return head, mapper
}

//the returned ast.node is the parent if in this case
func isElse(n ast.Node, file *ast.File, fset *token.FileSet) bool {
	interval, _ := astutil.PathEnclosingInterval(file, n.Pos(), n.End())
	if len(interval) < 2 { //|| exact {
		return false
	}
	//here we know that the node is not an if, and that it has at
	//least 1 parent if
	if ifn, ok := interval[1].(*ast.IfStmt); ok {
		if ifn.Else == n {
			fmt.Printf("Found The else \n%s\n", nodeToString(n))
			/*
				fmt.Printf("Found The else \n%s\n", nodeToString(n))
				var i = 1
				for ok {
					i++
					_, ok = interval[i].(*ast.IfStmt)
				}
				parent = interval[i]
			*/
			return true
		}
	}
	return false
}

func findParent(n ast.Node, head *Flow, file *ast.File, fset *token.FileSet) (*Flow, error) {
	interval, _ := astutil.PathEnclosingInterval(file, n.Pos(), n.End())
	if len(interval) < 2 { //|| exact {
		return nil, fmt.Errorf("Node has no parent in its ast")
	}
	fmt.Println(nodeToString(interval[1]))
	parentNode := interval[1]
	parentFlow := findFlowByNode(head, parentNode, fset)
	if parentFlow == nil {
		return nil, fmt.Errorf("Could not find flow's parent")
	}
	return parentFlow, nil
}

//depth first search of cfg
func findFlowByNode(f *Flow, n ast.Node, fset *token.FileSet) *Flow {
	//basecase
	fmt.Printf("Flowline %d : Nodeline %d", f.Line, fset.Position(n.Pos()).Line)
	if f.Line == fset.Position(n.Pos()).Line {
		return f
	}
	if len(f.Child) == 0 {
		return nil
	} else {
		for _, child := range f.Child {
			c := findFlowByNode(child, n, fset)
			if c != nil {
				return c
			}
		}
	}
	return nil
}

func nodeToString(n ast.Node) string {
	fset := token.NewFileSet()
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, n)
	return buf.String()
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
