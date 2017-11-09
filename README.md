[![GoDoc](https://godoc.org/github.com/fredericlemoine/golaxy?status.svg)](https://godoc.org/github.com/fredericlemoine/golaxy)
# golaxy : interacting with Galaxy in Go

[Golang](https://golang.org/) API to interact with the [Galaxy](https://usegalaxy.org/) API.

# Example of usage

## Launch a single tool and monitor its execution

```go
package main

import (
	"fmt"
	"time"

	"github.com/fredericlemoine/golaxy"
)

func main() {
	var err error
	var g *golaxy.Galaxy
	var historyid string
	var infileid string
	var jobids []string
	var outfiles map[string]string
	var jobstate string
	var filecontent []byte
	var tl *golaxy.ToolLaunch

	g = golaxy.NewGalaxy("http://galaxyip:port", "apikey", false)

	// Creates new history
	if historyid, err = g.CreateHistory("My history"); err != nil {
		panic(err)
	}

	// Uploads a file
	if infileid, _, err = g.UploadFile(historyid, "/path/to/file","auto"); err != nil {
		panic(err)
	}

	// Searches for the right tool id
	var my_tool string
	if tools, err := g.SearchToolID("my_tool"); err != nil {
		panic(err)
	} else {
		if len(tools)==0 {
			panic("no tool found")
		} else {
			fmt.Println(fmt.Sprintf("%d tools found",len(tools)))
			my_tool=tools[len(tools)-1]
		}
	}

	// Launches a new  Job
	tl = g.NewToolLauncher(historyid, my_tool)
	tl.AddParameter("option", "optionvalue")
	tl.AddFileInput("input", infileid, "hda")
	
	if _, jobids, err = g.LaunchTool(tl); err != nil {
		panic(err)
	}
	if len(jobids) < 1 {
		panic("No jobs")
	}

	end := false
	for !end {
		// Checks job status
		if jobstate, outfiles, err = g.CheckJob(jobids[0]); err != nil {
			panic(err)
		}

		if jobstate == "ok" {
			for _, id := range outfiles {
				// Downloads output files
				if filecontent, err = g.DownloadFile(historyid, id); err != nil {
					panic(err)
				}
				fmt.Println(string(filecontent))
			}
			end = true
		} else {
			fmt.Println("State:" + jobstate)
			for name, id := range outfiles {
				fmt.Println(name + "=>" + id)
			}
		}
		time.Sleep(2 * time.Second)
	}
	// Deletes history
	if jobstate, err = g.DeleteHistory(historyid); err != nil {
		panic(err)
	}
	fmt.Println(jobstate)
}
```

## Launch a workflow and monitor its execution

```go
package main

import (
	"fmt"
	"time"

	"github.com/fredericlemoine/golaxy"
)

func main() {
	g := golaxy.NewGalaxy("http://galaxyip:port", "apikey", false)
	var err error
	var historyid string
	var infileid string
	var wfids []string
	var wfinvocation *golaxy.WorkflowInvocation
	var workflowstate *golaxy.WorkflowStatus

	// Creates a new history
	if historyid, err = g.CreateHistory("My history"); err != nil {
		panic(err)
	}

	// Searches the workflow with given id (checks that it exists) 
	if wfids, err = g.SearchWorkflowIDs("cc1fde5b44055598"); err != nil {
		panic(err)
	}

	// Uploads input file
	if infileid, _, err = g.UploadFile(historyid, "/path/to/file", "auto"); err != nil {
		panic(err)
	}

	// Initializes a launcher
	l := g.NewWorkflowLauncher(historyid, wfids[0])
	l.AddFileInput("<input number>", infileid, "hda")
	l.AddParameter(<step number>, "param name", "param value")
	if wfinvocation, err = g.LaunchWorkflow(l); err != nil {
		panic(err)
	}

	// Now waits for the end of the execution
	end := false
	for !end {
		if workflowstate, err = g.CheckWorkflow(wfinvocation); err != nil {
			panic(err)
		}
		fmt.Println("Workflow Status: " + workflowstate.Status())
		// For all steps 
		for steprank := range workflowstate.ListStepRanks() {
			var status string
			var outfilenames []string
			var fileid string
			// Status of the given step
			if status, err = workflowstate.StepStatus(steprank); err == nil {
				fmt.Println("\t Job " + fmt.Sprintf("%d", steprank) + " = " + status)
				// Names and ids of outfiles of this step
				if outfilenames, err = workflowstate.StepOutFileNames(steprank); err != nil {
					panic(err)
				}
				for _, name := range outfilenames {
					if fileid, err = workflowstate.StepOutputFileId(steprank, name); err == nil {
						fmt.Println("\t\tFile " + name + " = " + fileid)
					}
				}
			}
		}
		end = (workflowstate.Status() == "ok" || workflowstate.Status() == "unknown" || workflowstate.Status()=="deleted")
		time.Sleep(2 * time.Second)
	}
	
	// Deletes history
	if jobstate, err = g.DeleteHistory(historyid); err != nil {
		panic(err)
	}

}
```
