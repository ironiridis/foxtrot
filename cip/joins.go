package cip

import "sync"

type DigitalTransition struct {
	Join  Join
	Value Digital
}

type AnalogTransition struct {
	Join  Join
	Value Analog
}

type SerialTransition struct {
	Join  Join
	Value Serial
}

func (d *DigitalTransition) Bytes() []byte {
	return nil
}

func (a *AnalogTransition) Bytes() []byte {
	return nil
}

func (s *SerialTransition) Bytes() []byte {
	return nil
}

type signalTransition interface {
	Bytes() []byte
}

// Joins reflects the current state of a CIP session.
type Joins struct {
	mu   sync.RWMutex
	din  map[Join]Digital
	ain  map[Join]Analog
	sin  map[Join]Serial
	dout map[Join]Digital
	aout map[Join]Analog
	sout map[Join]Serial

	out chan signalTransition
}

func (j *Joins) receive(rx chan signalTransition) {
	for r := range rx {
		switch t := r.(type) {
		case *DigitalTransition:
			j.DigitalIn(t.Join, t.Value)
		case *AnalogTransition:
			j.AnalogIn(t.Join, t.Value)
		case *SerialTransition:
			j.SerialIn(t.Join, t.Value)
		}
	}
}

func (j *Joins) send(tx signalTransition) {
	j.mu.RLock()
	if j.out != nil {
		j.out <- tx
	}
	j.mu.RUnlock()
}

func (j *Joins) DigitalIn(i Join, v Digital) {
	j.mu.Lock()
	if v {
		j.din[i] = v
	} else {
		delete(j.din, i)
	}
	j.mu.Unlock()
}

func (j *Joins) AnalogIn(i Join, v Analog) {
	j.mu.Lock()
	if v > 0 {
		j.ain[i] = v
	} else {
		delete(j.ain, i)
	}
	j.mu.Unlock()
}

func (j *Joins) SerialIn(i Join, v Serial) {
	j.mu.Lock()
	if len(v) > 0 {
		j.sin[i] = make([]byte, len(v))
		copy(j.sin[i], v)
	} else {
		delete(j.sin, i)
	}
	j.mu.Unlock()
}

func (j *Joins) SetDigital(i Join, v Digital) {
	j.mu.Lock()
	if v {
		j.dout[i] = v
	} else {
		delete(j.dout, i)
	}
	j.mu.Unlock()
	j.send(&DigitalTransition{Join: i, Value: v})
}

func (j *Joins) SetAnalog(i Join, v Analog) {
	j.mu.Lock()
	if v > 0 {
		j.aout[i] = v
	} else {
		delete(j.aout, i)
	}
	j.mu.Unlock()
	j.send(&AnalogTransition{Join: i, Value: v})
}

func (j *Joins) SetSerial(i Join, v Serial) {
	j.mu.Lock()
	if len(v) > 0 {
		j.sout[i] = make([]byte, len(v))
		copy(j.sout[i], v)
	} else {
		delete(j.sout, i)
	}
	j.mu.Unlock()
	j.send(&SerialTransition{Join: i, Value: v})
}

func (j *Joins) outChan(ch chan signalTransition) {
	j.mu.Lock()
	j.out = ch
	j.mu.Unlock()
}

func (j *Joins) Sync() {
	j.mu.RLock()
	for i := range j.dout {
		j.send(&DigitalTransition{Join: i, Value: j.dout[i]})
	}
	for i := range j.aout {
		j.send(&AnalogTransition{Join: i, Value: j.aout[i]})
	}
	for i := range j.sout {
		j.send(&SerialTransition{Join: i, Value: j.sout[i]})
	}
	j.mu.RUnlock()
}
