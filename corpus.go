package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// idp, np, and lp are patterns for recognizing metadata fields
var idp = regexp.MustCompile("ID=([0-9]+)")
var np = regexp.MustCompile("NAME=\"([^\"]+)\"")
var lp = regexp.MustCompile("LANGUAGE=\"([A-Z]+)\"")

// SpeakerMeta holds the meta information about a speaker.
type SpeakerMeta struct {
	Language string `json:"language"`
	Name     string `json:"name"`
	ID       string `json:"id"`
}

type searchStats struct {
	started   time.Time
	processed int
	resets    int
	missing   int
}

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

// findSpeakerMeta moves upwards within a source transcript and searches for speaker meta information.
func findSpeakerMeta(source []string, line int) (sm SpeakerMeta) {
	if line == 0 {
		return sm
	}

	// move upwards from current line until the beginning
	for i := line - 1; i >= 0; i-- {
		// if a chapter comes first, abort - this is some meta information about the session.
		if strings.HasPrefix(source[i], "<CHAPTER") {
			return sm
		} else if strings.HasPrefix(source[i], "<SPEAKER") {
			m := lp.FindStringSubmatch(source[i])
			if len(m) == 2 {
				sm.Language = m[1]
			}
			m = idp.FindStringSubmatch(source[i])
			if len(m) == 2 {
				sm.ID = m[1]
			}
			m = np.FindStringSubmatch(source[i])
			if len(m) == 2 {
				sm.Name = m[1]
			}

			return sm
		}
	}
	return sm
}

func findSourceSentence(sentence string, sourceSentences [][]string, i int, j int) (int, int) {
	for k := i; k < len(sourceSentences); k++ {
		for l := j; l < len(sourceSentences[k]); l++ {
			if strings.Contains(sourceSentences[k][l], sentence) {
				return k, l
			}
		}
	}
	return -1, -1
}

func findSentences(
	wg *sync.WaitGroup,
	id int,
	alignments chan Alignment,
	missing chan string,
	a []string,
	b []string,
	sourceSentences [][]string,
	offset int,
) {
	defer wg.Done()

	j := 0
	k := 0

	s := searchStats{
		started:   time.Now(),
		processed: 0,
		resets:    0,
		missing:   0,
	}

	for i, sentence := range a {
		// some aligned sentences may started with spaces, protected whitespaces (U+00A0), periods, and others
		trimmedSentence := strings.Trim(sentence, "  .,:;-–—()")
		l, m := findSourceSentence(trimmedSentence, sourceSentences, j, k)
		if l == -1 || m == -1 {
			// we tried our best to find it within the regular order, now do a full search re-run
			l, m = findSourceSentence(trimmedSentence, sourceSentences, 0, 0)
			if l == -1 || m == -1 {
				missing <- sentence
				s.missing++
				continue
			}
			s.resets++
		}
		j = l
		k = m

		sm := findSpeakerMeta(sourceSentences[j], k)

		// FIXME language IDs
		alignments <- Alignment{
			Speaker: sm,
			Sentences: []Sentence{
				{Language: "de", Sentence: sentence},
				{Language: "en", Sentence: b[offset+i]},
			},
		}
		s.processed++

		if i > 0 && i%25000 == 0 {
			sooner := time.Now()
			diff := sooner.Sub(s.started)
			log.Printf("🎱 %d Processed %d sentences, %d resets (%d/s)\n", id, i, s.resets, int(math.RoundToEven(float64(s.processed)/diff.Seconds())))

			s.started = sooner
			s.processed = 0
			s.missing = 0
			s.resets = 0
		}
	}
	log.Printf("🏁 %d Worker finished.\n", id)
}
