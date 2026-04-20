package sst

const (
	SequenceNoMatch             = "sequence_not_found"
	SequenceStageNoMatch        = "sequence_stage_not_found"
	SequenceMilestoneNotReached = "sequence_milestone_unreachable"
	StageGatesPassed            = "stage_gate_passing"
	StageGatesNotPassing        = "stage_gate_not_passing"
)

type Stage struct {
	Name string

	PropertyGates  []*PropertyGateDefinition
	ComponentGates []*ComponentGateDefinition
	LogicGates     []*LogicGateDefinition

	position int
	next     *Stage
}

type Sequence struct {
	stages     map[string]*Stage // internal lookup only
	stageSlice []*Stage
}

func (s *Sequence) AddStage(stage *Stage) *Sequence {
	if len(s.stageSlice) == 0 {
		stage.position = 1
	} else {
		last := s.stageSlice[len(s.stages)-1]
		stage.position = last.position + 1
		last.next = stage
	}
	s.stages[stage.Name] = stage
	s.stageSlice = append(s.stageSlice, stage)
	return s
}

type Orchestration struct {
	sequences map[string]*Sequence // internal lookup only
}

func NewOrchestration() *Orchestration {
	return &Orchestration{
		sequences: map[string]*Sequence{},
	}
}

func (o *Orchestration) SequenceFor(kind string) *Sequence {
	s := &Sequence{
		stages: map[string]*Stage{},
	}
	o.sequences[kind] = s
	return s
}

func (o *Orchestration) ValidateTransition(e Entity, milestoneStage string) (bool, []*Message) {
	var seq *Sequence
	var ok bool
	if seq, ok = o.sequences[e.Kind()]; !ok {
		return false, []*Message{
			{
				Content: SequenceNoMatch,
				ID:      e.ID(),
				Kind:    e.Kind(),
			},
		}
	}

	var stage *Stage
	if stage, ok = seq.stages[e.Stage()]; !ok {
		return false, []*Message{
			{
				Content: SequenceStageNoMatch,
				ID:      e.ID(),
				Kind:    e.Kind(),
				Context: map[string]string{
					"stage": e.Stage(),
				},
			},
		}
	}

	var milestone *Stage
	if milestone, ok = seq.stages[milestoneStage]; !ok {
		return false, []*Message{
			{
				Content: SequenceStageNoMatch,
				ID:      e.ID(),
				Kind:    e.Kind(),
				Context: map[string]string{
					"milestone": milestoneStage,
				},
			},
		}
	}

	success := true
	var messages []*Message

	for stage != milestone {
		next := stage.next
		if next == nil {
			break
		}

		succeedsStage := true
		lgs := map[string]*LogicGate{}
		var logicGates []*LogicGate
		for _, lg := range next.LogicGates {
			// we store it redundantly as we need it as lookup and loop
			gate := lg.Gate()
			lgs[lg.name] = gate
			logicGates = append(logicGates, gate)
		}

		for _, pg := range next.PropertyGates {
			pass, issues := pg.Evaluate(e.Properties(), e.Kind(), e.ID())
			messages = append(messages, issues...) // we always record all messages, as it is not responsibility to handle them; just to record

			// either it is reference on a logic gate, then we defer the flagging of issues
			// or it ain't that it is a direct reason to fail
			var lg *LogicGate
			if lg, ok = lgs[pg.logicGate]; ok {
				lg.Input(pass)
			}

			// if output is not redirected to logic gate and is false
			if !pass && lg == nil {
				succeedsStage = false
				success = false
			}
		}

		for _, cg := range next.ComponentGates {
			/**
			note:
			as of now the preferred stage must be defined by the user, i.e. the stage that the system tries to
			transition an entity to. Though there is an implication that the system can determine the stage, as either
			the furthest or earliest possible one from the list provided and therefore could auto-pick the preferred
			stage. This might be a future long hanging fruit to simplify the interface.
			*/
			bc := cg.BufferedComponent()
			for _, c := range e.Components() {
				componentStage := c.Stage()

				if !bc.Passes(c) {
					ps := bc.PreferredStage()
					// if the current stage of component is not accepted; try if a transition to the preferred stage
					// is possible; if so count it as a pass, and pass along the transition
					// messages
					if ps != "" {
						pass, issues := o.ValidateTransition(c, ps)
						messages = append(messages, issues...)
						if pass {
							componentStage = ps
						}
					}
				}

				// feed into the buffer component gate component; to allow later evaluation based on "valid" count
				bc.Input(c.Kind(), componentStage)
			}

			pass, issues := bc.Evaluate(e.Kind(), e.ID())
			messages = append(messages, issues...)

			// either it is reference on a logic gate, then we defer the flagging of issues
			// or it ain't that it is a direct reason to fail
			var lg *LogicGate
			if lg, ok = lgs[bc.LogicGate()]; ok {
				lg.Input(pass)
			}

			// if output is not redirected to logic gate and is false
			if !pass && lg == nil {
				succeedsStage = false
				success = false
			}
		}

		for _, g := range logicGates {
			if !g.Evaluate() {
				messages = append(messages, &Message{
					Content: LogicGateDidNotPass,
					ID:      e.ID(),
					Kind:    e.Kind(),
					Context: map[string]string{
						"gate":      g.Name(),
						"milestone": next.Name,
					},
				})
				succeedsStage = false
				success = false
			}
		}

		if succeedsStage {
			messages = append(messages, &Message{
				Content: StageGatesPassed,
				ID:      e.ID(),
				Kind:    e.Kind(),
				Context: map[string]string{
					"from_stage": stage.Name,
					"to_stage":   next.Name,
				},
			})
		} else {
			messages = append(messages, &Message{
				Content: StageGatesNotPassing,
				ID:      e.ID(),
				Kind:    e.Kind(),
				Context: map[string]string{
					"from_stage": stage.Name,
					"to_stage":   next.Name,
				},
			})
		}

		stage = next
	}

	if stage != milestone {
		success = false
		messages = append(messages, &Message{
			Content: SequenceMilestoneNotReached,
			ID:      e.ID(),
			Kind:    e.Kind(),
			Context: map[string]string{
				"milestone": milestoneStage,
			},
		})
	}

	return success, messages
}
