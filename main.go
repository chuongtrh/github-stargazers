package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/chuongtrh/github-stargazers/stargazers"
	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Need github repo to process")
	}

	repoStr := args[0]
	fmt.Println(repoStr)

	rand.Seed(time.Now().Unix())

	soup.Headers = map[string]string{
		"user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36",
	}

	stargazers, err := stargazers.GetStargazers(repoStr)
	if err != nil {
		log.Fatal(err)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Username", "Follower", "Contribution this year"})

	temp := stargazers[0:int(math.Min(100, float64(len(stargazers))))]

	for i, u := range temp {
		t.AppendRow([]interface{}{i + 1, u.Username, u.Follower, u.Contribution})
	}
	t.Render()
}
