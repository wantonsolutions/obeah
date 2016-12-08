package obeah

import (
	"encoding/gob"
	"log"
	"os"
)

var (
	initalized = false
	targets    map[string]map[string]map[string]Target
	sTargets   map[string]Target
	traces     [][]Target
)

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
	for _, t := range tr {
		logger.Println(t.Id)
	}
    logger.Println()
	//start a new trace
	traces = append(traces, make([]Target, 0))
}

func checkInit() {
	if !initalized {
		initNow()
	}
}
