package obeah

import (
	"encoding/gob"
	"log"
	"os"
    "fmt"
)

const (
    BOUND = 5
)

var (
	initalized = false
	targets    map[string]map[string]map[string]Target
	sTargets   map[string]Target
	traces     [][]Target

    //Conditional Graph
    heads = make(map[string]*Node,0)
    nodes = make(map[string]*Node,0)
)

type Node struct {
    Hits int
    Tar Target
    Children map[string]*Node
    ChildrenHits map[string]int
}

func (n *Node) String() string {
    var children string
    for _, c := range n.Children {
        children += c.Tar.Id + "\n"
    }
    return fmt.Sprintf("Hits: %d, Target:%s, Children: %s",n.Hits,n.Tar,children)
}

func NewNode() *Node {
    return &Node{Hits: 0, Tar: NewTarget(), Children: make(map[string]*Node,0), ChildrenHits: make(map[string]int)}
}

func pathVariables(n []*Node) map[string]Variable {
    vars := make(map[string]Variable,0)
    for i := range n {
        for j := range n[i].Tar.Vars {
            vars[j] = n[i].Tar.Vars[j]
        }
    }
    return vars
}

func getPathCondition(n []*Node) string {
    if len(n) <= 0 {
        return ""
    }
    var pathCondition string
    conditions := make([]string,0)
    for i := range n {
        for j := range n[i].Tar.Condition {
            conditions = append(conditions,n[i].Tar.Condition[j])
        }
    }
    for i := 0; i < len(conditions)-1 ;i++ {
        pathCondition += fmt.Sprintf("%s && ",conditions[i])
    }
    pathCondition += fmt.Sprintf("%s",conditions[len(conditions)-1])
    return pathCondition
}

func generatePath() []*Node {
    min := 1.0
    index := 0
    ps := make([][]string,0)
    for i := range heads {
        ps = append(ps,dfs(heads[i],BOUND,make([]string,0))...)
    }
    for i := range ps {
        s := getScore(ps[i])
        if s < min {
            min = s
            index = i
        }
    }
    return fetchNodes(ps[index])
}

func fetchNodes(path []string) []*Node {
    n := make([]*Node,0)
    for _, id := range path {
        n = append(n,nodes[id])
    }
    return n
}

func getScore(p []string) float64 {
    score := 1.0
    for i :=0; i < len(p)-1;i++ {
        total := 0 
        for j := range nodes[p[i]].ChildrenHits {
            total += nodes[p[i]].ChildrenHits[j]
        }
        score *= float64(nodes[p[i]].ChildrenHits[p[i+1]]) / float64(total)
    }
    return score
}

func dfs(n *Node, bound int, path []string) [][]string {
    if len(n.Children) == 0 || bound == 0 {
        return append(make([][]string,0),path)
    }
    paths := make([][]string,0)
    for i := range n.Children {
        paths = append(paths,dfs(n.Children[i],bound-1,append(path,n.Children[i].Tar.Id))...)
    }
    return paths
}

func processTrace(tr []Target) {
    //catch base case
    if len(tr) <= 0 {
        return
    }
    //check trace head exits
    if _, ok := heads[tr[0].Id]; !ok {
        //check if head has been seen before
        if _, ok = nodes[tr[0].Id]; !ok {
        //this is a new node and a new head
            n := NewNode()
            n.Tar = tr[0]
            nodes[n.Tar.Id] = n
        }
        //append the new head to the list
        heads[tr[0].Id] = nodes[tr[0].Id]
    }
    //heads are added, but have not been updated
    //Process the trace here adding new children as they are found
    //stop one short of the end of the trace
    for i:= 0 ;i < len(tr) -1; i++ {
        nodes[tr[i].Id].Hits++
        //update child
        //check if the child node exists
        if _, ok := nodes[tr[i+1].Id]; !ok {
            //child has not yet been seen
            cn := NewNode()
            cn.Tar = tr[i+1]
            nodes[cn.Tar.Id] = cn
        }
        //child now exists update node
        nodes[tr[i].Id].Children[tr[i+1].Id] = nodes[tr[i+1].Id]
        nodes[tr[i].Id].ChildrenHits[tr[i+1].Id]++
    }
    //all nodes in the trace are in the map, last has not been accounted for
    //now update hits for the final node in the trace (must exist!)
    nodes[tr[len(tr)-1].Id].Hits++
}
    

func initNow() {
	if initalized {
		return
	}

	logger = log.New(os.Stdout, "[Obeah Runtime] ", log.Lshortfile)
	//read in target file
	file, err := os.Open("targets.enc")
	if err != nil {
		logger.Fatal("Unable to Init Obeah: %s", err)
	}
	file.Seek(0, 0)
	gob.Register(Target{})
	dec := gob.NewDecoder(file)
	err = dec.Decode(&targets)
	if err != nil {
		logger.Fatalf("Unable to Init Obeah: %s", err.Error())
	}
	//TODO done while this is in the testing phase to work at the
	//packe level use more than just sTargets
    sTargets = make(map[string]Target,5)
	for p := range targets {
		for s := range targets[p] {
            for k, v := range targets[p][s] {
			    sTargets[k] = v
            }
		}
	}

	//TODO get a better estimate for the length of traces and try to
	//read old ones in for continued testing
	traces = make([][]Target, 1)
	initalized = true
}

func Log(id string, names string, vars ...interface{}) {
	checkInit()
	traces[len(traces)-1] = append(traces[len(traces)-1], sTargets[id])
}

//Taboo messes up your program
func Taboo(id, names string, vars ...interface{}) {
	checkInit()
	//print the last trace
	tr := traces[len(traces)-1]
    processTrace(tr)
    if len(traces) > 20 && len(traces) % 5 == 0 {
        path := generatePath()
        pv := pathVariables(path)
        cont := getPathCondition(path)
        
        for _, v := range pv {
            logger.Println(v.String())
        }
        logger.Println(cont)
        DrawDot(heads)
    }
	//start a new trace
	traces = append(traces, make([]Target, 0))
}

func checkInit() {
	if !initalized {
		initNow()
	}
}

//functions for drawing the runtime cfg
func DrawDot(map[string]*Node) {
    var entries []string
    visited := make(map[*Node]bool)
    for i := range heads {
        traverseAndDraw(heads[i],&entries,visited)
    }
    file, err := os.Create("runtimeCFG.dot")
    if err != nil {
        logger.Printf("Unable to write runtime CFG: %s",err.Error())
        return
    }
    file.WriteString("digraph runtimeCFG {\n")
    for _, e := range entries {
        file.WriteString(fmt.Sprintf("\t%s\n", e))
    }
    file.WriteString("}\n")
}

func traverseAndDraw(n *Node, entries *[]string,visited map[*Node]bool){
    if visited[n] {
        return
    }
    visited[n] = true
    for i := range n.Children {
        connection := fmt.Sprintf("%s -> %s [ label = \"%d\" ] ",n.Tar.Id, n.Children[i].Tar.Id,n.ChildrenHits[i])
        *entries = append(*entries,connection)
        traverseAndDraw(n.Children[i],entries,visited)
    }
    return
}
