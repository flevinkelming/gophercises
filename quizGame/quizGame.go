package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type problem struct {
	Q, A string
}

type problems []problem

func main() {
	csvFileName := flag.String("csv", "problems.csv", "A .csv file in the format of"+
		"'question,answer'")
	timeLimit := flag.Int("limit", 30, "The time allotted for the quiz in seconds.")
	flag.Parse()

	f, err := os.Open(*csvFileName)
	if err != nil {
		log.Fatalf("Failed to load .csv file: %s\n", *csvFileName)
	}

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("An error occurred while parsing the csv: %s\n", err.Error())
	}

	quiz := parseCSVData(rows)

	correct := 0
	timer := time.NewTimer(time.Duration(*timeLimit) * time.Second)
	for i, p := range quiz {
		fmt.Printf("Question #%d: %s? ", i+1, p.Q)

		answerChan := make(chan string)
		go func() {
			var answer string
			fmt.Scanf("%s", &answer)
			answerChan <- answer
		}()

		select {
		case <-timer.C:
			fmt.Printf("\nYou answered %d/%d questions correctly.", correct, len(quiz))
			return
		case answer := <-answerChan:
			if answer == p.A {
				correct++
			}
		}
	}

	fmt.Printf("You answered %d/%d questions correctly.", correct, len(quiz))
}

func parseCSVData(rows [][]string) problems {
	quiz := make(problems, len(rows))
	for i, row := range rows {
		quiz[i] = problem{
			Q: strings.TrimSpace(row[0]),
			A: strings.TrimSpace(row[1]),
		}
	}

	return quiz
}
