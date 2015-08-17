/*
Library for Chinese vocabulary analysis
*/
package analysis

import (
	"bufio"
	"container/list"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

// Defines the sense of a Chinese word
type WordSenseEntry struct {
	Id int
	Simplified, Traditional, Pinyin, English, Grammar, Concept_cn,
			Concept_en, Topic_cn, Topic_en, Parent_cn, Parent_en, Image,
			Mp3, Notes string
}

// The dictionary is a map of pointers to word senses, indexed by simplified
// and traditional text
var wdict map[string][]*WordSenseEntry

// Breaks text into a list of CJK and non CJK strings
func GetChunks(text string) (list.List) {
	var chunks list.List
	cjk := ""
	noncjk := ""
	for _, character := range text {
		if IsCJKWord(string(character)) {
			if noncjk != "" {
				chunks.PushBack(noncjk)
				noncjk = ""
			} 
			cjk += string(character)
		} else if cjk != "" {
			chunks.PushBack(cjk)
			cjk = ""
			noncjk += string(character)
		} else {
			noncjk += string(character)
		}
	}
	if cjk != "" {
		chunks.PushBack(cjk)
	}
	if noncjk != "" {
		chunks.PushBack(noncjk)
	}
	return chunks
}

// Gets the dictionary definition of the first word sense matching the word
func GetWordSense(chinese string) (WordSenseEntry, bool) {
	wSenses, ok := wdict[chinese]
	if ok {
		return *(wSenses[0]), ok
	}
	ws := new(WordSenseEntry)
	return *ws, ok
}

// Gets the dictionary definition of a word
// Parameters
//   chinese: The Chinese (simplified or traditional) text of the word
// Return
//   word: an array of word senses
//   ok: true if the word is in the dictionary
func GetWord(chinese string) (word []*WordSenseEntry, ok bool) {
	word, ok = wdict[chinese]
	return word, ok
}

func IsCJKWord(character string) bool {
	_, ok := wdict[character]
	return ok
}

// Reads the Chinese-English dictionary into memory from the word sense file
// Parameters:
//   wsfilename The name of the word sense file
func ReadDict(wsfilename string) {
	wsfile, err := os.Open(wsfilename)
	if err != nil {
		log.Fatal(err)
	}
	defer wsfile.Close()
	reader := csv.NewReader(wsfile)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	wdict = make(map[string][]*WordSenseEntry)
	for _, row := range rawCSVdata {
		id, _ := strconv.ParseInt(row[0], 10, 0)
		simp := row[1]
		trad := row[2]
		newWs := &WordSenseEntry{Id: int(id), Simplified: simp,
				Traditional: trad, Pinyin: row[3], English: row[4],
				Grammar: row[5], Concept_cn: row[6], Concept_en: row[7], 
				Topic_cn: row[8], Topic_en: row[9], Parent_cn: row[10],
				Parent_en: row[11], Image: row[12], Mp3: row[13],
				Notes: row[14]}
		if trad != "\\N" {
			wSenses, ok := wdict[trad]
			if !ok {
				wsSlice := make([]*WordSenseEntry, 1)
				wsSlice[0] = newWs
				wdict[trad] = wsSlice
			} else {
				wdict[trad] = append(wSenses, newWs)
			}
		}
		wSenses, ok := wdict[simp]
		if !ok {
			wsSlice := make([]*WordSenseEntry, 1)
			wsSlice[0] = newWs
			wdict[simp] = wsSlice
		} else {
			//fmt.Printf("ReadDict: found simplified %s already in dict\n", simp)
			wdict[simp] = append(wSenses, newWs)
		}
	}
}

// Reads a Chinese text file
func ReadText(filename string) (string) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	text := string(bs)
	//fmt.Printf("ReadText: read text %s\n", text)
	return text
}

// Parses a Chinese text into words
// text: the string to parse
// tokens: the tokens for the parsed text
// vocab: a table of the unique words found in the parsed text
func ParseText(text string) (tokens list.List, vocab map[string]bool) {
	vocab = make(map[string]bool)
	chunks := GetChunks(text)
	//fmt.Printf("ParseText: For text %s got %d chunks\n", text, chunks.Len())
	for e := chunks.Front(); e != nil; e = e.Next() {
		chunk := e.Value.(string)
		//fmt.Printf("ParseText: chunk %s\n", chunk)
		characters := strings.Split(chunk, "")
		if !IsCJKWord(characters[0]) {
			tokens.PushBack(chunk)
			continue
		}
		for i := 0; i < len(characters); i++ {
			for j := len(characters); j > 0; j-- {
				w := strings.Join(characters[i:j], "")
				//fmt.Printf("ParseText: i = %d, j = %d, w = %s\n", i, j, w)
				if _, ok := wdict[w]; ok {
					tokens.PushBack(w)
					vocab[w] = true
					i += j - 1
					j = 0
					//fmt.Printf("ParseText: found word %s, i = %d\n", w, i)
				}
			}
		}
	}
	return tokens, vocab
}

// Writes a document with markup for the array of tokens
// tokens: A list of tokens forming the document
// vocab: A list of word id's in the document
// filename: The file name to write to
func WriteDoc(tokens list.List, vocab map[string]bool, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for e := tokens.Front(); e != nil; e=e.Next() {
		word := e.Value.(string)
		//fmt.Printf("WriteDoc: Word %s\n", word)
		if entries, ok := GetWord(word); ok {
			wordIds := ""
			for _, ws := range entries {
				if wordIds == "" {
					wordIds = fmt.Sprintf("%d", ws.Id)
				} else {
					wordIds = fmt.Sprintf("%s,%d", wordIds, ws.Id)
				}
			}
			fmt.Fprintf(w, "<span title='%s' data-wordid='%s'" +
					" class='dict-entry' data-toggle='popover'>%s</span>",
					word, wordIds, word)
		} else {
			index := strings.Index(word, "<!-- words here -->")
			if index == -1 {
				w.WriteString(word)
			} else {
				w.WriteString(word[:index])
				w.WriteString("\n")
				w.WriteString("<script>\n")
				w.WriteString("words = {\n")
				for key, _ := range vocab {
					if entries, ok := GetWord(key); ok {
						for _, ws := range entries {
							fmt.Fprintf(w, "\"%d\":{\"element_text\":\"%s\"," +
									"\"simplified\":\"%s\"," +
									"\"traditional\":\"%s\"," +
									"\"pinyin\":\"%s\",\"english\":\"%s\"," +
									"\"notes\":\"%s\"},\n", ws.Id, key,
									ws.Simplified, ws.Traditional, ws.Pinyin,
									ws.English, ws.Notes)
						}
					} 
				}
				w.WriteString("}\n")
				w.WriteString("</script>\n")
				w.WriteString(word[index:])
				w.WriteString("\n")
			}
		}
	}
	w.Flush()
}
