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
	"golang.org/x/tools/go/loader"
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

//Target is a structure for holding control flow information about a given conditional
//The name Target is meant to imply that these will be amied for at runtime.
type Target struct {
	Id        string
	Line      int
	Condition []string
	Vars      map[string]Variable
//	Node      ast.Node
}

type Variable struct {
	Id    string
	Name  string
	Type  string
	Value interface{}
}

type Brackets struct {
    L, R []token.Pos
}

func NewBrackets() Brackets {
    return Brackets{L: make([]token.Pos,0), R: make([]token.Pos,0)}
}

func (b Brackets) String() string{
    var bs string
    bs = "{ -> ["
    for _, l := range b.L {
        bs += fmt.Sprintf("%d,",l)
    }
    bs += "],\t } -> ["
    for _, r := range b.R {
        bs += fmt.Sprintf("%d,",r)
    }
    bs += "]"
    return bs
}
    

func (b Brackets) depth (pos token.Pos) int {
    d := 0
    for _, lb := range b.L {
        if lb < pos {
            d++
        }
    }
    for _, rb := range b.R {
        if rb < pos {
            d--
        }
    }
    return d
}





func NewTarget() Target {
	return Target{Id: "", Line: -1, Condition: make([]string, 0), Vars: make(map[string]Variable, 0)}
}

func NewVariable() Variable {
	return Variable{Id: "", Name: "", Type: "", Value: nil}
}


func (t Target) String() string {
	var vars string
	for key := range t.Vars {
		vars += t.Vars[key].String() + "\n"
	}
	return fmt.Sprintf("Id:%s Line:%d Condition:%s Vars[%s]", t.Id, t.Line, condToString(t.Condition), vars)
}

func (v Variable) String() string {
	return fmt.Sprintf("Id: %s\tName: %s\tType: %s\tValue:%s", v.Id, v.Name, v.Type, v.Value)
}

func Insturment(options map[string]string, l *log.Logger) (map[string]string, map[string]map[string]map[string]Target) {
	logger = l
	logger.SetPrefix("[Obeah Instrument] ")
	initalize(options)
	p, err := getProgramWrapper()
	if err != nil {
		panic(err)
	}
	instrumentedOutput := make(map[string]string)
	targets := make(map[string]map[string]map[string]Target, 0)
	for pnum, pack := range p.Packages {
		targets[pack.PackageName] = make(map[string]map[string]Target, 0)
		for snum, soc := range pack.Sources {
			var instSource string
			instSource, targets[pack.PackageName][soc.Filename] = InstrumentSource(p.Fset, p.Packages[pnum].Sources[snum].Source, p.Prog)
			p.Packages[pnum].Sources[snum].Text = instSource
			instrumentedOutput[p.Packages[pnum].Sources[snum].Filename] = instSource
		}
	}

	return instrumentedOutput, targets
}

func InstrumentSource(fset *token.FileSet, file *ast.File, p *loader.Program) (string, map[string]Target) {
	lines := ControlFlowLines(fset, file, p)
	buf := new(bytes.Buffer)
	printer.Fprint(buf, fset, file)
	split := strings.SplitAfter(buf.String(), "\n")
	mergedSource := make([]string, 0)
	id := 0
	//ast.Print(fset, file)
	for i := range split {
		mergedSource = append(mergedSource, split[i])
		if _, ok := lines[i+1]; ok {
			//mergedSource = append(mergedSource,fmt.Sprintf("obeah.Log(`%d`)\n",i+1))
			cond := condToString(lines[i+1].Condition)
			mergedSource = append(mergedSource, "obeah.Log(\""+lines[i+1].Id+"\",\""+cond+"\")\n")
			id++
		}
	}
	referencedTargets := make(map[string]Target, 0)
	for _, t := range lines {
		referencedTargets[t.Id] = t
	}
	instrumented := mergeSource(mergedSource)
	//fmt.Println(instrumented)
	formatted, err := format.Source([]byte(instrumented))
	if err != nil {
		panic(err)
	}
	return string(formatted), referencedTargets
}

func getBrackets(fset *token.FileSet, file *ast.File) Brackets {
    b := NewBrackets()
	ast.Inspect(file, func(n ast.Node) bool {
        switch s := n.(type) {
        case *ast.BlockStmt:
            b.L = append(b.L,s.Lbrace)
            b.R = append(b.R,s.Rbrace)
            break
        case *ast.CompositeLit:
            b.L = append(b.L,s.Lbrace)
            b.R = append(b.R,s.Rbrace)
            break
        default:
            return true
        }
        return true
        })
    return b
}

func ControlFlowLines(fset *token.FileSet, file *ast.File, p *loader.Program) map[int]Target {
	mapper := make(map[int]Target)
    brackets := getBrackets(fset,file)
    logger.Println(brackets.String())
	ast.Inspect(file, func(n ast.Node) bool {
		switch c := n.(type) {
		//function entrance
		case *ast.FuncDecl:
			t := NewTarget()
			t.Id = fmt.Sprintf("%d",fset.Position(c.Body.Pos()).Offset)
			t.Line = fset.Position(c.Body.Pos()).Line
			mapper[t.Line] = t
			//if statement //must come before
		case *ast.IfStmt:
			t := NewTarget()
			t.Id = fmt.Sprintf("%d",fset.Position(c.Body.Pos()).Offset)
			t.Line = fset.Position(c.Body.Pos()).Line
			//get parent conditions
			con, vars := getCondition(n, brackets ,file, fset, p)
			t.Vars = vars
			t.Condition = append(t.Condition, con...)
			//get local conditions
			t.Condition = append(t.Condition, nodeToString(c.Cond))
			getVarsFromCond(c.Cond, file, p, t.Vars)
			logger.Println(t.String())
			mapper[t.Line] = t
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
	return mapper
}

//variables is filled with the variables from the conditional
func getVarsFromCond(c ast.Node, file *ast.File, p *loader.Program, variables map[string]Variable) {
	defs := p.Created[0].Defs
	ast.Inspect(c, func(n ast.Node) bool {
		switch i := n.(type) {
		case *ast.Ident:
			for d := range defs {
				if i.Obj == d.Obj { //the objects match
					obj := defs[d]
					v := NewVariable()
					v.Name = obj.Name()
					v.Id = obj.Id()
					v.Type = obj.Type().String()
					variables[v.Id] = v
				}
			}
			break
		default:
			break
		}
		return true
	})
}

func getCondition(n ast.Node, b Brackets,file *ast.File, fset *token.FileSet, p *loader.Program) ([]string, map[string]Variable) {
	interval, _ := astutil.PathEnclosingInterval(file, n.Pos(), n.End())
	condition := make([]string, 0)
	variables := make(map[string]Variable, 0)
	for i := 1; i < len(interval); i++ {
		switch c := interval[i].(type) {
		case *ast.IfStmt:
            //if they are on the same level negate
            if b.depth(n.Pos()) == b.depth(c.Pos()) {
			    condition = append(condition, "!("+nodeToString(c.Cond)+")")
            } else {
			    condition = append(condition, "("+nodeToString(c.Cond)+")")
            }
			getVarsFromCond(c.Cond, file, p, variables)
			break
		default:
			break
		}
	}
	return condition, variables
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

func condToString(cond []string) string {
	if len(cond) <= 0 {
		return ""
	}
	var ret string
	for i := 0; i < len(cond)-1; i++ {
		//fmt.Println(cond[i])
		ret += cond[i] + " && "
	}
	ret += cond[len(cond)-1]
	return ret
}

//merges b into a
func mapmerge(a, b map[string]Variable) {
	for key := range b {
		a[key] = b[key]
	}
}
