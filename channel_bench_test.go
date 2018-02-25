package channel_test

import (
	"testing"
	"github.com/x12tech/channel"
	"sync"
)

type emptyArg struct{}


func BenchmarkChanCom(b *testing.B) {
	proc := channel.NewProcessor(`some`,100)
	proc.AddHandler(func(e *emptyArg){})
	proc.Start()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		proc.Input(&emptyArg{})
	}
}


func BenchmarkPlainChanCom(b *testing.B) {
	ch := make(chan *emptyArg,100)
	go func() {
		for {
			select {
			case s  := <-ch:
				_ = s
			}
		}
	}()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ch <- &emptyArg{}
	}
}

type respArg struct {
	resp chan struct{}
}

func BenchmarkChanComResponce(b *testing.B) {
	proc := channel.NewProcessor(`some`,100)
	p := sync.Pool{
		New: func() interface{} {
			return  &respArg{
				resp: make(chan struct{},1),
			}
		},
	}
	proc.AddHandler(func(e *respArg){
		e.resp <- struct{}{}
	})
	proc.Start()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		a := p.Get().(*respArg)
		proc.Input(a)
		<-a.resp
		p.Put(a)
	}
}


func BenchmarkPlainChanComResponce(b *testing.B) {
	ch := make(chan *respArg,100)
	p := sync.Pool{
		New: func() interface{} {
			return  &respArg{
				resp: make(chan struct{},1),
			}
		},
	}
	go func() {
		for {
			select {
			case s  := <-ch:
				s.resp <- struct{}{}
			}
		}
	}()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		a := p.Get().(*respArg)
		ch <- a
		<-a.resp
		p.Put(a)
	}
}
