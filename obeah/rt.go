package obeah

import (
	"encoding/gob"
	"log"
	"os"
    "fmt"
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
        children += c.String() + "\n"
    }
    return fmt.Sprintf("Hits: %d, Target:%s, Children: %s",n.Hits,n.Tar,children)
}

func NewNode() *Node {
    return &Node{Hits: 0, Tar: NewTarget(), Children: make(map[string]*Node,0), ChildrenHits: make(map[string]int)}
}


var (
	initalized = false
	targets    map[string]map[string]map[string]Target
	sTargets   map[string]Target
	traces     [][]Target

    //Conditional Graph
    heads = make(map[string]*Node,0)
    nodes = make(map[string]*Node,0)
)

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

    for k, v := range sTargets {
        logger.Printf("target :%s key: %s\n",v.String(),k)
    }
	//TODO get a better estimate for the length of traces and try to
	//read old ones in for continued testing
	traces = make([][]Target, 1)
	initalized = true
}

func Log(id string, extra ...interface{}) {
	checkInit()
	traces[len(traces)-1] = append(traces[len(traces)-1], sTargets[id])
}

//Taboo messes up your program
func Taboo(vars ...interface{}) {
	checkInit()
	//print the last trace
	tr := traces[len(traces)-1]
    processTrace(tr)
    if len(traces) % 5 == 0 {
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
        connection := fmt.Sprintf("%s -> %s",n.Tar.Id, n.Children[i].Tar.Id)
        *entries = append(*entries,connection)
        traverseAndDraw(n.Children[i],entries,visited)
    }
    return
}
