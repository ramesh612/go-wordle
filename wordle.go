package main

// takes an optional word list as a command line argument. if none provided, tries to use
// the word lists from standard linux locations like /usr/share/dict list is 'cleansed'
// and only 'valid' 5 letter words are loaded into memory.
// once the above steps happen the engine is ready to accept connections and play
// usual wordle rules apply. total 6 attempts. there are 2 modes - hard and easy.
// feedback provided after every valid guess. invalid words are not accepted as guesses
// if unable to solve after 6 attempts the word is displayed at the end.
// engine also saves every session into a in memory database and dumps the data from db
// to file system upon shutdown and/or periodically.

import (
	"bufio"
	"fmt"
	gc "github.com/gbin/goncurses"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func in(words []string, word string) bool {
	for _, v := range words {
		if v == word {
			return true
		}
	}
	return false
}

func isValid(word string, words []string) bool {
	return len(word) == 5 && in(words, word)
}

func initializeWords() []string {
	var validWords []string
	dictName := "/usr/share/dict/words"

	file, err := os.Open(dictName)
	handleErr(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		if word[0] >= 65 && word[0] <= 90 {
			// skip words that begin with a capital letter (likely proper noun)
		} else if strings.ContainsAny(word, "'") {
			// skip words that contain apostrophe
		} else if len(word) != 5 {
			// skip words that are not wordle length
		} else {
			validWords = append(validWords, word)
		}
	}
	log.Printf("loaded %d words into memory from %s", len(validWords), dictName)
	return validWords
}

func contains(rs []rune, r rune) bool {
	for _, val := range rs {
		if val == r {
			return true
		}
	}
	return false
}

func cowsAndBulls(guess string, pick string) ([]rune, []rune) {
	var locMatch, inWord []rune
	i := 0
	for i < 5 {
		if guess[i] == pick[i] {
			// location Matched
			log.Printf("locMatch: %c\n", rune(guess[i]))
			locMatch = append(locMatch, rune(guess[i]))
		} else if strings.ContainsRune(pick, rune(guess[i])) {
			log.Printf("inWord: %c\n", rune(guess[i]))
			inWord = append(inWord, rune(guess[i]))
		}
		i++
	}
	var nonLocMatch []rune
	for _, val := range inWord {
		if !contains(locMatch, val) {
			nonLocMatch = append(nonLocMatch, val)
		}
	}
	return locMatch, nonLocMatch
}

func pickRandomWord(validWords []string) string {
	rand.Seed(time.Now().UnixNano())
	return validWords[rand.Intn(len(validWords))]
}

func main() {
	stdscr, err := gc.Init()
	handleErr(err)
	defer gc.End()

	f, err := os.OpenFile("wordle.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	handleErr(err)
	defer f.Close()

	log.SetOutput(f)

	validWords := initializeWords()
	pickedWord := pickRandomWord(validWords)
	log.Printf("picked %s\n", pickedWord)
	guessCount := 0
	found := false
	for guessCount < 6 {
		msg := "please enter your guess #" + strconv.Itoa(guessCount+1) + ": "
		stdscr.MovePrint(0, 0, msg)

		var str string
		str, err = stdscr.GetString(5)
		if err != nil {
			stdscr.MovePrint(guessCount, 0, "GetString Error:", err)
		} else {
			if !isValid(str, validWords) {
				log.Printf("%s is not a valid word\n", str)
				stdscr.MovePrintf(guessCount, 0, "You entered: %s. It is not a valid wordle word", str)
			} else if str == pickedWord {
				guessCount++
				log.Printf("found in %d tries\n", guessCount)
				stdscr.MovePrintf(guessCount, 0, "You entered: %s. %s", str, "congratulations, you found the word!")
				found = true
				break
			} else {
				guessCount++
				locMatch, inWord := cowsAndBulls(str, pickedWord)
				log.Printf("locaMatch: %v, inWord: %v\n", locMatch, inWord)
				msg := fmt.Sprintf("location match: [%s], in the word: [%s]", string(locMatch), string(inWord))
				stdscr.MovePrintf(guessCount, 0, "You entered: %s. %s", str, msg)
			}
		}
	}
	if guessCount == 6 && !found {
		row := guessCount + 1
		stdscr.MovePrintf(row, 0, "You have used up all six attempts. Word is %s", pickedWord)
	}
	stdscr.GetChar()
}
