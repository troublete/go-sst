package sst

const (
	LogicGateDidNotPass = "logic_gate_not_passing"
)

type LogicGateDefinition struct {
	name string
	isOr bool
}

func And(name string) *LogicGateDefinition {
	return &LogicGateDefinition{
		name: name,
	}
}

func Or(name string) *LogicGateDefinition {
	return &LogicGateDefinition{
		name: name,
		isOr: true,
	}
}

type LogicGate struct {
	lgd      *LogicGateDefinition
	inputs   [2]bool // simplified true and false inputs
	hasInput bool
}

func (lgd *LogicGateDefinition) Gate() *LogicGate {
	return &LogicGate{
		lgd: lgd,
	}
}

func (lg *LogicGate) Input(value bool) {
	lg.hasInput = true
	if value {
		lg.inputs[1] = true
	} else {
		lg.inputs[0] = true
	}
}

func (lg *LogicGate) Evaluate() bool {
	if !lg.hasInput {
		return false
	}

	// if AND logic gate and only true exists; if OR logic gate and at least any one true exists
	if (lg.lgd.isOr && lg.inputs[1]) || (!lg.lgd.isOr && !lg.inputs[0]) {
		return true
	}

	return false
}

func (lg *LogicGate) Name() string {
	return lg.lgd.name
}
