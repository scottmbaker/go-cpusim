package cpusim

type EnablerInterface interface {
	Bool() bool
}

type TrueEnabler struct {
}

type FalseEnabler struct {
}

type ByReferenceEnabler struct {
	Value    *bool
	Inverted bool
}

type EnableBit struct {
	Value    bool
	LoEnable ByReferenceEnabler
	HiEnable ByReferenceEnabler
}

func (e *TrueEnabler) Bool() bool {
	return true
}

func (e *FalseEnabler) Bool() bool {
	return false
}

func (e *ByReferenceEnabler) Bool() bool {
	if e.Inverted {
		return !(*e.Value)
	} else {
		return *e.Value
	}
}

func (e *EnableBit) Set(value bool) {
	e.Value = value
}

func NewEnableBit() *EnableBit {
	eb := &EnableBit{}
	eb.LoEnable = ByReferenceEnabler{Value: &eb.Value, Inverted: true}
	eb.HiEnable = ByReferenceEnabler{Value: &eb.Value, Inverted: false}
	return eb
}

var AlwaysEnabled TrueEnabler
var AlwaysDisabled FalseEnabler
