package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for dataRaw := range in {
		var data string
		switch v := dataRaw.(type) {
		case int:
			data = strconv.Itoa(v)
		case string:
			data = v
		default:
			return
		}

		md5 := DataSignerMd5(data)
		wg.Add(1)
		go func(data string, md5 string) {
			defer wg.Done()

			chDataCrc := make(chan string)
			chMd5Crc := make(chan string)

			go func() {
				chDataCrc <- DataSignerCrc32(data)
				close(chDataCrc)
			}()
			go func(md5 string) {
				chMd5Crc <- DataSignerCrc32(md5)
				close(chMd5Crc)
			}(md5)

			result := <-chDataCrc + "~" + <-chMd5Crc
			out <- result
		}(data, md5)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for dataRaw := range in {
		wg.Add(1)
		go func(dataRaw interface{}) {
			defer wg.Done()
			data, ok := dataRaw.(string)
			if !ok {
				return
			}

			res := make([]string, 6)
			var innerWG sync.WaitGroup

			for i := 0; i < 6; i++ {
				innerWG.Add(1)
				go func(i int) {
					defer innerWG.Done()
					res[i] = DataSignerCrc32(strconv.Itoa(i) + data)
				}(i)
			}

			innerWG.Wait()
			out <- strings.Join(res, "")
		}(dataRaw)
	}

	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	datas := make([]string, 0)

	for dataRaw := range in {
		data, ok := dataRaw.(string)
		if !ok {
			continue
		}

		datas = append(datas, data)
	}

	sort.Strings(datas)
	out <- strings.Join(datas, "_")
}

func ExecutePipeline(jobs ...job) {
	chans := make([]chan interface{}, len(jobs)+1)
	for i := range chans {
		chans[i] = make(chan interface{})
	}

	close(chans[0])

	var wg sync.WaitGroup
	wg.Add(len(jobs))
	for i := 0; i < len(jobs); i++ {
		go func(i int, in, out chan interface{}) {
			defer wg.Done()
			jobs[i](in, out)
			close(out)
		}(i, chans[i], chans[i+1])
	}

	wg.Wait()
}


// go run 2/99_hw/signer/signer.go 2/99_hw/signer/common.go
// func main() {
// 	chans := make([]chan interface{}, 4)
// 	for i := range chans {
// 		chans[i] = make(chan interface{})
// 	}

// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		SingleHash(chans[0], chans[1])
// 		close(chans[1])
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		MultiHash(chans[1], chans[2])
// 		close(chans[2])
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		CombineResults(chans[2], chans[3])
// 		close(chans[3])
// 	}()

// // если бы функции сами закрывали каналы
// // go SingleHash(chans[0], chans[1])
// // go MultiHash(chans[1], chans[2])
// // go CombineResults(chans[2], chans[3])

// 	chans[0] <- "0"
// 	chans[0] <- "1"
// 	close(chans[0])

// 	go func() {
// 		wg.Wait()
// 	}()

// 	for res := range chans[3] {
//         fmt.Printf("Get result: %s\n", res)
//     }
// }
