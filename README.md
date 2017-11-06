[![GoDoc](https://godoc.org/github.com/fredericlemoine/golaxy?status.svg)](https://godoc.org/github.com/fredericlemoine/golaxy)
# golaxy : interacting with Galaxy in Go

A first implementation of basic functions to call the [Galaxy](https://usegalaxy.org/) API in [Golang](https://golang.org/).

# Example of usage
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

	g = golaxy.NewGalaxy("http://galaxyip:port", "apikey", false)

	/* Create new history */
	if historyid, err = g.CreateHistory("My history"); err != nil {
		panic(err)
	}

	/* Upload a file */
	if infileid, _, err = g.UploadFile(historyid, "/path/to/file","auto"); err != nil {
		panic(err)
	}

	/* Searching the right tool id */
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

	/* Launch Job */
	mapfiles := make(map[string]string)
	mapfiles["input"] = infileid
	params := make(map[string]string)
	params["option"] = "optionvalue"
	if _, jobids, err = g.LaunchTool(historyid, my_tool, mapfiles, params); err != nil {
		panic(err)
	}
	if len(jobids) < 1 {
		panic("No jobs")
	}

	end := false
	for !end {
		/* Check job state */
		if jobstate, outfiles, err = g.CheckJob(jobids[0]); err != nil {
			panic(err)
		}

		if jobstate == "ok" {
			for _, id := range outfiles {
				/* Download output files */
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
	/* Delete history */
	if jobstate, err = g.DeleteHistory(historyid); err != nil {
		panic(err)
	}
	fmt.Println(jobstate)
}
```
