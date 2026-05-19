/*
	WAL - Write Ahead Log. Writes down function calls in case server crashes.
*/

package store

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type WAL struct {
	file *os.File
}

func NewWAL(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return &WAL{file: file}, nil
}

func (w *WAL) Append(op string, key string, value string) error {
	_, err := fmt.Fprintf(w.file, "%s %s %s\n", op, key, value)
	if err != nil {
		return err
	}
	return w.file.Sync()
}

func Replay(path string) ([][]string, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	results := [][]string{}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		results = append(results, parts)
	}

	return results, nil
}
