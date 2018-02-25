package channel_test

import (
	"time"
	"github.com/x12tech/channel"


	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("channel.Processor", func() {
	It("works", func(done Done) {
		var count int
		str := `test`
		proc := channel.NewProcessor(`test`, 100)
		proc.AddHandler(func(value *string) {
			defer GinkgoRecover()
			Expect(value).To(Equal(&str))
			count++
		})
		proc.Start()

		Expect(proc.Input(&str)).To(Succeed())
		Expect(proc.Input(&str)).To(Succeed())
		proc.Stop()
		Expect(count).To(Equal(2))
		close(done)
	}, 1)
	It("timer works", func(done Done) {
		var count int
		str := `test`
		proc := channel.NewProcessor(`test`, 100)
		proc.AddHandler(func(value *string) {
			defer GinkgoRecover()
			Expect(value).To(Equal(&str))
			count++
		})
		fire := make(chan struct{}, 100)
		proc.AddTimer(time.Nanosecond, func() {
			fire <- struct{}{}
		})
		fire2 := make(chan struct{}, 100)
		proc.AddTimer(time.Nanosecond, func() {
			fire2 <- struct{}{}
		})
		proc.Start()
		<-fire
		<-fire2
		proc.Stop()
		close(done)
	}, 1)
	It("handler works", func(done Done) {
		type cbType struct {
			test int
		}

		var count int
		str := `test`
		proc := channel.NewProcessor(`test`, 100)
		proc.AddHandler(func(value *string) {
			defer GinkgoRecover()
			Expect(value).To(Equal(&str))
			count++
		})
		proc.AddHandler(func(value *cbType) {
			defer GinkgoRecover()
			Expect(value.test).To(Equal(1))
			count++
		})
		proc.Start()

		Expect(proc.Input(&str)).To(Succeed())
		Expect(proc.Input(&str)).To(Succeed())
		Expect(proc.Input(&cbType{
			test: 1,
		})).To(Succeed())
		proc.Stop()
		Expect(count).To(Equal(3))
		close(done)
	}, 1)
})

type TestStorage1 struct {
	dataArray []*string
	proc      *channel.Processor
}

func NewTestStorage1() *TestStorage1 {
	self := &TestStorage1{}
	proc := channel.NewProcessor(`test`, 100)
	proc.AddHandler(self.OnMessage)
	proc.AddTimer(time.Second*10, self.flushBuf)
	proc.OnFinalize(self.flushBuf)
	self.proc = proc
	self.proc.Start()
	return self
}

func (self *TestStorage1) Stop() {
	self.proc.Stop()
}

func (self *TestStorage1) Store(str *string) {
	self.proc.Input(str)
}

func (self *TestStorage1) OnMessage(value *string) {
	self.dataArray = append(self.dataArray, value)
}

func (self *TestStorage1) flushBuf() {
	//write self.dataArray someWhere
}
