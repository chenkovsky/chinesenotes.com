/*
Library for Chinese vocabulary analysis
Word frequency functions.
*/
package analysis

import (
	"sort"
	"fmt"
	)

// A word with corpus entry label
type CorpusWord struct {
	Corpus, Word string
}

// A word frequency with corpus entry label
type CorpusWordFreq struct {
	Corpus, Word string
	Freq int
}

// Sorted list of word frequencies
type SortedWF struct {
	wf map[string]int
	w []SortedWordItem
}

// An entry in a sorted word array
type SortedWordItem struct {
	Word string
	Freq int
}

// For indexing counts
func (cw CorpusWord) String() string {
	return fmt.Sprintf("%s:%s", cw.Corpus, cw.Word)
}

func (sortedWF *SortedWF) Len() int {
	return len(sortedWF.wf)
}

func (sortedWF *SortedWF) Less(i, j int) bool {
	return sortedWF.wf[sortedWF.w[i].Word] > sortedWF.wf[sortedWF.w[j].Word]
}

func (sortedWF *SortedWF) Swap(i, j int) {
	sortedWF.w[i], sortedWF.w[j] = sortedWF.w[j], sortedWF.w[i]
}

/*
 * Sorts Word struct's based on frequency
 */
func SortedFreq(wf map[string]int) []SortedWordItem {
	sortedWF := new(SortedWF)
	sortedWF.wf = wf
	sortedWF.w = make([]SortedWordItem, len(wf))
	i := 0
	for key, _ := range wf {
		sortedWF.w[i] = SortedWordItem{key, sortedWF.wf[key]}
		i++
	}
	sort.Sort(sortedWF)
	return sortedWF.w
}