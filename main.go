package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Batches states how many batches we do process in parallel
const Batches = 4

// <https://stackoverflow.com/a/18479916>
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func readSources(path string) ([][]string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	sources := make([][]string, len(files))
	for i, f := range files {
		if f.IsDir() {
			continue
		}

		lines, err := readLines(filepath.Join(path, f.Name()))
		if err != nil {
			return nil, err
		}
		sources[i] = lines
	}
	return sources, nil
}

func findSentences(wg *sync.WaitGroup, id int, sentences []string, sources [][]string, offset int) {
	defer wg.Done()

	for i, sentence := range sentences {
		found := false

		// some aligned sentences may start with spaces, protected whitespaces (U+00A0), and periods
		sentence = strings.Trim(sentence, "  .")

		for _, sourceFile := range sources {
			for _, sourceSentence := range sourceFile {
				if strings.Contains(sourceSentence, sentence) {
					found = true
				}
			}
		}

		if !found {
			log.Printf("🚨 ️Could not find sentence of line %d\nSentence: %s\nPrefix runes: %U\nSuffix runes: %U\n\n", offset+i, sentence, []rune(sentence[:8]), []rune(sentence[(len(sentence)-8):]))
		}

		if i%1000 == 0 && i > 0 {
			log.Printf("Worker %d worked %d sentences.\n", id, i)
		}
	}

	log.Printf("🏁 Worker %d finished.\n", id)
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: $ eurol1 <aligned-l1> <aligned-l2> <source-l1>")
		os.Exit(0)
	}

	aligned, err := readLines(os.Args[1])
	if err != nil {
		panic(err)
	}
	log.Println("Aligned files read.")

	sources, err := readSources(os.Args[3])
	if err != nil {
		panic(err)
	}
	log.Println("Source files read.")

	batchSize := int(math.Ceil(float64(len(aligned)) / float64(Batches)))
	log.Printf("Input (aligned sentences): %d\n", len(aligned))
	log.Printf("Batch size: %d\n", batchSize)

	var wg sync.WaitGroup
	wg.Add(Batches)

	for i := 0; i < Batches; i++ {
		start := i * batchSize
		end := start + batchSize - 1
		batch := aligned[start:end]

		log.Printf("Starting batch %d (%d to %d)\n", i, start, end)
		go findSentences(&wg, i, batch, sources, start)
	}

	wg.Wait()
	log.Println("Done.")
}
