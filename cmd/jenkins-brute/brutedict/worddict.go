// modified from github.com/dieyushi/golang-brutedict
package brutedict

import "fmt"

type WordDict struct {
	start   int
	end     int
	queue   chan []string
	quit    chan bool
	running bool
}

func NewWordDict(dict []string, start, end int) (bd *WordDict) {
	bd = &WordDict{
		start:   start,
		end:     end,
		running: true,
		queue:   make(chan []string),
		quit:    make(chan bool),
	}

	var b = make([]string, end)
	go bd.process(dict, b, start, end)
	return
}

func (bc *WordDict) CombinationsChan() chan []string {
	return bc.queue
}

func (bd *WordDict) process(dict []string, b []string, start int, end int) {
	defer func() { recover() }()

	for i := start; i <= end; i++ {
		fmt.Printf("Permuting dict strings of word length %v\n", i)
		bd.list(dict, b, i, 0)
	}
	close(bd.queue)
	bd.quit <- true
}

func (bd *WordDict) Id() (str []string) {
	select {
	case str = <-bd.queue:
	case <-bd.quit:
	}
	return
}

func (bd *WordDict) Close() {
	bd.running = false
	close(bd.queue)
}

func (bd *WordDict) list(dict []string, b []string, l int, j int) {
	strl := len(dict)

	for i := 0; i < strl; i++ {
		b[j] = dict[i]
		if j+1 < l {
			bd.list(dict, b, l, j+1)
		} else {
			strCopy := make([]string, len(b))
			copy(strCopy, b)
			bd.queue <- strCopy
		}
	}
}
