package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	strCh := make(chan string)

	go func () {
		line := ""
		for {
			buf := make([]byte, 8)
			n, err := f.Read(buf)
			if err != nil && err != io.EOF {
				return
			}

			chunk := string(buf[:n])
			if strings.Contains(chunk, "\n") {
				lastPart := ""
				parts := strings.Split(chunk, "\n")
				for i, part := range parts {
					if i != len(parts) - 1  {
						line += part
					} else {
						lastPart = part
					}
					
				}
				strCh <- line
				line = lastPart
			} else {
				line += chunk
			}


			if err == io.EOF {
				if line != "" {
					strCh <- line
				}
				close(strCh)
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