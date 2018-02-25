package channel

import (
	"errors"
	"reflect"
	"sync"
	"time"
)

var str = ``
var StringPtr = reflect.TypeOf(&str)
var String = reflect.TypeOf(str)

type ticker struct {
	t        *time.Ticker
	cb       func()
	duration time.Duration
}

type Processor struct {
	inputCh      chan reflect.Value
	timersCh     chan int
	exitCh       chan struct{}
	tickers      map[int]*ticker
	tickersIdCnt int
	started      bool
	onFinalize   func()
	exitWg       sync.WaitGroup
	name         string
	handlers     map[reflect.Type]reflect.Value
}


func NewProcessor(name string, queueLen int) *Processor {
	return &Processor{
		inputCh:  make(chan reflect.Value, queueLen),
		timersCh: make(chan int, 10),
		exitCh:   make(chan struct{}, 1),
		tickers:  make(map[int]*ticker),
		handlers: make(map[reflect.Type]reflect.Value),
		name:     name,
	}
}

func (proc *Processor) Start() {
	proc.started = true
	proc.exitWg.Add(1)
	proc.startTimers()
	go proc.runLoop()
}

func (proc *Processor) Stop() {
	if proc.started {
		proc.exitCh <- struct{}{}
		proc.exitWg.Wait()
	}
}

func (proc *Processor) startTimers() {
	for id, v := range proc.tickers {
		tick := time.NewTicker(v.duration)
		v.t = tick
		go func(timerId int) {
			for range tick.C {
				proc.timersCh <- timerId
			}
		}(id)
	}
}

func (proc *Processor) stopTimers() {
	for _, v := range proc.tickers {
		v.t.Stop()
	}
}

func (proc *Processor) AddTimer(duration time.Duration, cb func()) error {
	if proc.started {
		return errors.New(`Processor already started`)
	}
	newId := proc.tickersIdCnt
	proc.tickersIdCnt++
	proc.tickers[newId] = &ticker{
		cb:       cb,
		duration: duration,
	}
	return nil
}

func (proc *Processor) AddHandler(cb interface{}) error {
	handle := reflect.ValueOf(cb)
	typ := handle.Type()
	if typ.Kind() != reflect.Func {
		return errors.New(`cb must be a function`)
	}
	if typ.NumIn() != 1 {
		return errors.New(`cb must be a function with 1 argument`)
	}
	inType := typ.In(0)

	proc.handlers[inType] = handle
	return nil
}

func (proc *Processor) OnFinalize(cb func()) {
	proc.onFinalize = cb
}

func (proc *Processor) runLoop() {
LOOP:
	for {
		select {
		case value := <-proc.inputCh:
			proc.handleInput(value)
		case timerId := <-proc.timersCh:
			timer := proc.tickers[timerId]
			if timer != nil {
				timer.cb()
			} else {
				panic(`internal logic error: timer does not exists`)
			}
		case <-proc.exitCh:
			break LOOP
		}
	}
	proc.finalize()
}

func (proc *Processor) finalize() {
	proc.stopTimers()
	// empty input queue
LOOP:
	for {
		select {
		case value := <-proc.inputCh:
			proc.handleInput(value)
		default:
			break LOOP
		}
	}
	if proc.onFinalize != nil {
		proc.onFinalize()
	}
	proc.exitWg.Done()
}

func (proc *Processor) handleInput(value reflect.Value) {
	typ := value.Type()
	cb, ok := proc.handlers[typ]
	if ok {
		cb.Call([]reflect.Value{ value })
	} else {
		panic(`internal logic error: handler for type ` + typ.String() + ` does not exists`)
	}
}

func (proc *Processor) Input(value interface{})  error {
	v := reflect.ValueOf(value)
	typ := v.Type()

	_, ok := proc.handlers[typ]
	if !ok {
	 return  errors.New(`internal logic error: handler for type ` + typ.String() + ` does not exists`)
	}
	proc.inputCh <- v
	return nil
}
