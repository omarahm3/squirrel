package main

import (
	"bufio"
	"log"
	"os"
)

func main() {
  input := make(chan string)

  InitClient(input)
}

func ScanFile(input chan string) {
  scanner := bufio.NewScanner(os.Stdin)
  for scanner.Scan() {
    text := scanner.Text()
    input <- text
  }

  if err := scanner.Err(); err != nil {
    log.Println(err)
  }
}
