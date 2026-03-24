package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	strCh := make(chan string)

	go func () {
		defer close(strCh)
		line := ""
		for {
			chunk := make([]byte, 8)
			n, err := f.Read(chunk)
			if err != nil && err != io.EOF {
				break
			}

			chunk = chunk[:n]
			if i := bytes.IndexByte(chunk, '\n'); i != -1 {
				line += string(chunk[:i])
				chunk = chunk[i+1:]
				strCh <- line
				line = ""
			}
			line += string(chunk)

			if err == io.EOF {
				if line != "" {
					strCh <- line
				}
				break
			}
		}

	}()

	return strCh
}

func main() {
	f, err := os.Open("message.txt")
	if err != nil {
		return
	}

	defer f.Close()

	
	for str := range getLinesChannel(f) {
		fmt.Printf("read: %s\n", str)
	}

}