package json

type pipeline struct {
	final        []byte
	err          error
	pendingBytes []byte
	queue        chan queueItem
	done         chan struct{}
}

type queueItem struct {
	pr   <-chan promiseResult
	bs   []byte
	term bool
}

type promiseResult struct {
	bytes []byte
	err   error
}

func newPipeline() *pipeline {
	pl := &pipeline{
		queue: make(chan queueItem, 512),
		done:  make(chan struct{}),
	}
	go pl.run()
	return pl
}

func (p *pipeline) run() {
	for qi := range p.queue {
		if qi.term {
			close(p.done)
		} else if qi.pr != nil {
			res := <-qi.pr
			if res.err != nil {
				p.err = res.err
			}
			p.final = append(p.final, res.bytes...)
		} else {
			p.final = append(p.final, qi.bs...)
		}
	}
}

func (p *pipeline) appendBytes(bs []byte) {
	p.pendingBytes = append(p.pendingBytes, bs...)
}

func (p *pipeline) appendByte(b byte) {
	p.pendingBytes = append(p.pendingBytes, b)
}

func (p *pipeline) appendPromise(ch <-chan promiseResult) {
	p.flushPendingBytes()
	p.queue <- queueItem{pr: ch}
}

func (p *pipeline) flush() ([]byte, error) {
	p.flushPendingBytes()
	p.queue <- queueItem{term: true}
	<-p.done
	close(p.queue)
	return p.final, p.err
}

func (p *pipeline) flushPendingBytes() {
	if len(p.pendingBytes) > 0 {
		p.queue <- queueItem{bs: p.pendingBytes}
		p.pendingBytes = nil
	}
}
