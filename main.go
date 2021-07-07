package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sync"
)

// Batches states how many batches we do process in parallel
const Batches = 8

type Alignment struct {
	Speaker   SpeakerMeta `json:"speaker"`
	Sentences []Sentence  `json:"sentences"`
}

type Sentence struct {
	Language string `json:"language"`
	Sentence string `json:"sentence"`
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: $ eurol1 <aligned-l1> <aligned-l2> <source-l1> out.json")
		os.Exit(0)
	}

	l1, err := readLines(os.Args[1])
	if err != nil {
		panic(err)
	}
	l2, err := readLines(os.Args[2])
	if err != nil {
		panic(err)
	}
	sources, err := readSources(os.Args[3])
	if err != nil {
		panic(err)
	}
	outfile := os.Args[4]

	log.Println("Input files read.")

	batchSize := int(math.Ceil(float64(len(l1)) / float64(Batches)))
	log.Printf("Input (aligned sentences): %d\n", len(l1))
	log.Printf("Batch size: %d\n", batchSize)

	var wg sync.WaitGroup
	wg.Add(Batches)

	alignments := make(chan Alignment, len(l1))
	missing := make(chan string, len(l1))

	go func() {
		for i := 0; i < Batches; i++ {
			start := i * batchSize
			end := min(start+batchSize-1, len(l1))
			batch := l1[start:end]

			log.Printf("Starting batch %d (%d to %d)\n", i, start, end)
			go findSentences(&wg, i, alignments, missing, batch, l2, sources, start)
		}
	}()

	wg.Wait()

	close(alignments)
	close(missing)

	log.Printf("%d sentences were missing in total.\n", len(missing))
	log.Printf("Writing files output...\n")

	var as []Alignment
	for a := range alignments {
		as = append(as, a)
	}
	d, _ := json.Marshal(as)
	_ = ioutil.WriteFile(outfile, d, 0644)

	var ms []string
	for m := range missing {
		if m != "" {
			ms = append(ms, m)
		}
	}
	d, _ = json.Marshal(ms)
	_ = ioutil.WriteFile("missing.json", d, 0644)

	log.Println("Done.")
}
