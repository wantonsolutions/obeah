package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/wantonsolutions/obeah/obeah"
	"log"
	"os"
	"regexp"
	"runtime/pprof"
)

const (
	//instrumenter defaults
	defaultFilename  = ""
	defaultDirectory = ""
	defaultPipe      = ""
)

var (
	//options for detecting
	directory string
	file      string
	pipe      string

	//options for both
	verbose bool
	debug   bool

	logger *log.Logger
)

func setFlags() {
	flag.StringVar(&directory, "dir", defaultDirectory, "-dir=directoryName recursivly instruments a directory inplace, original directory is duplicated for safty")
	flag.StringVar(&file, "file", defaultFilename, "-file=filename insturments a file")

	flag.BoolVar(&verbose, "verbose", false, "-verbose logs extensive output")
	flag.BoolVar(&verbose, "v", false, "-v logs extensive output")
	flag.BoolVar(&debug, "debug", false, "-debug adds pedantic level of logging")
	flag.Parse()
}

func main() {
	setFlags()

	options := make(map[string]string)
	//set options relevent to all programs
	if verbose {
		logger = log.New(os.Stdout, "[Obeah Setup] ", log.Lshortfile)
	} else {
		var buf bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	}

	if debug {
		options["debug"] = "on"
	}

	//filechecking //exclusive or with filename and directory
	if file == defaultFilename && directory == defaultDirectory {
		if len(os.Args) == 2 && !verbose {
			file = os.Args[1]
		} else {
			//try to read from pipe
			reader := bufio.NewReader(os.Stdin)
			// Read all data from stdin, processing subsequent reads as chunks.
			data := make([]byte, 100000) // Read 4MB at a time
			n, err := reader.Read(data)
			if err != nil {
				logger.Fatalf("Problems reading from input: %s", err)
			}
			buffer := bytes.NewBuffer(data)
			pipe += buffer.String()[0:n]
		}
	} else if file != defaultFilename && directory != defaultDirectory {
		logger.Fatalf("Speficied filename =%s and directory = %s, use either -file or -dir\n", file, directory)
	}

	if pipe != defaultPipe {
		options["pipe"] = pipe
		//TODO write targets to file
		source, _ := obeah.Insturment(options, logger)
		fmt.Print(source)
		return
	}

	//test if file exists, if so add file option
	if file != defaultFilename {
		exists, err := fileExists(file)
		if !exists {
			a := err.Error()
			print(a)
			logger.Fatalf("Error: : %s\n", err.Error())
		}
		logger.Printf("Documenting %s\n", file)

		options["file"] = file
		//get source
		source, targets := obeah.Insturment(options, logger)
		targetFile, err := os.Create("targets.enc")
		if err != nil {
			logger.Fatal(err)
		}
		enc := gob.NewEncoder(targetFile)
		enc.Encode(targets)
		printSource(source[file])
		err = writeFile(file, source[file])
		if err != nil {
			log.Fatal(err)
		}
	}

	// TODO remove test if the directory is valid. If so add to options, else
	// error
	if directory != defaultDirectory {
		valid, err := validDir(directory)
		if !valid {
			logger.Fatalf("Invalid Directory Error: %s\n", err.Error())
		}
		logger.Printf("Documenting Directory :%s\n", directory)
		//TODO write targets to file
		options["directory"] = directory

		sources, _ := obeah.Insturment(options, logger)
		for name, source := range sources {
			logger.Printf("%s\n%s\n", name, source)
			/*
				err := writeFile(name, source)
				if err != nil {
					log.Fatal(err)
				}
			*/
		}
	}

}

func printSource(source string) {
	fmt.Println(source)
}

func writeFile(filename, source string) error {
	//overwrite file
	ofile, err := os.OpenFile(filename, os.O_RDWR, os.FileMode(0666)) // For read access.
	defer ofile.Close()
	if err != nil {
		return err
	}
	err = ofile.Truncate(0)
	if err != nil {
		return err
	}
	logger.Printf("Writing over source of %s\n", filename)
	_, err = ofile.WriteString(source)
	if err != nil {
		return err
	}
	return nil
}

func validDir(dir string) (bool, error) {
	//TODO check that dir exists
	//TODO check for existing go args
	/*if len(args) != 3 {
		return false, fmt.Errorf("Directory or package non existant\n")
	}*/
	return true, nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//getCallingFunctionID returns the file name and line number of the
//program which called capture.go. This function is used to generate
//logging statements dynamically.
func getCallingFunctionID() string {
	profiles := pprof.Profiles()
	block := profiles[1]
	var buf bytes.Buffer
	block.WriteTo(&buf, 1)
	//fmt.Printf("%s",buf)
	passedFrontOnStack := false
	re := regexp.MustCompile("([a-zA-Z0-9]+.go:[0-9]+)")
	ownFilename := regexp.MustCompile("capture.go") // hardcoded own filename
	matches := re.FindAllString(fmt.Sprintf("%s", buf), -1)
	for _, match := range matches {
		if passedFrontOnStack && !ownFilename.MatchString(match) {
			return match
		} else if ownFilename.MatchString(match) {
			passedFrontOnStack = true
		}
		//fmt.Printf("found %s\n", match)
	}
	fmt.Printf("%s\n", buf)
	return ""
}
