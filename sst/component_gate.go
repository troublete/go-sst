package sst

const (
	ComponentGateCountTooLow  = "component_quantity_count_too_low"
	ComponentGateCountNoMatch = "component_quantity_count_no_match"
)

type ComponentGateDefinition struct {
	kind   string
	stages []string

	// if this is not empty, the system will try to move the found component to this stage if not already
	// in a supported stage already;
	preferredStage string

	stagesLookup map[string]int

	n, min    int
	logicGate string
}

func NewComponentGate(kind string, stages []string, preferredStage string, requiredN, minN int, gate ...string) *ComponentGateDefinition {
	cgd := &ComponentGateDefinition{
		kind:           kind,
		stages:         stages,
		preferredStage: preferredStage,
		n:              requiredN,
		min:            minN,
		stagesLookup:   map[string]int{},
	}

	// if a lookup here is required is debatable, for a small n of stages (~10) the lookup is actually more expensive
	// but to have it a little more approachable we create some memory overhead here
	for _, s := range stages {
		cgd.stagesLookup[s] = 0
	}

	if len(gate) > 0 {
		cgd.logicGate = gate[0]
	}

	return cgd
}

type ComponentGate struct {
	cgd *ComponentGateDefinition

	matchN, noMatchN int
}

func (cgd *ComponentGateDefinition) BufferedComponent() *ComponentGate {
	return &ComponentGate{
		cgd: cgd,
	}
}

func (cg *ComponentGate) Passes(e Entity) bool {
	if e.Kind() != cg.cgd.kind {
		return true // ignore not matching kind
	}

	if _, ok := cg.cgd.stagesLookup[e.Stage()]; ok {
		return true
	}

	return false
}

func (cg *ComponentGate) Input(kind, stage string) {
	if kind != cg.cgd.kind {
		return
	}

	if _, ok := cg.cgd.stagesLookup[stage]; ok {
		cg.matchN++
	} else {
		cg.noMatchN++
	}
}

func (cg *ComponentGate) PreferredStage() string {
	return cg.cgd.preferredStage
}

func (cg *ComponentGate) LogicGate() string {
	return cg.cgd.logicGate
}

func (cg *ComponentGate) Evaluate(postfix string) (bool, []string) {
	if cg.matchN+cg.noMatchN < cg.cgd.min {
		return false, []string{ComponentGateCountTooLow + ",component_kind=" + cg.cgd.kind + "," + postfix}
	}

	if cg.matchN < cg.cgd.n {
		return false, []string{ComponentGateCountNoMatch + ",component_kind=" + cg.cgd.kind + "," + postfix}
	}

	return true, nil
}
