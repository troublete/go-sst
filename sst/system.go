package sst

import (
	"context"
	"fmt"
)

type Gate struct {
	EntityKind string
	EntityID   string

	FromStage string
	ToStage   string
	Passes    bool

	Reason          string
	ReasonReference string
}

type ComponentGate struct {
	Kind      string
	StageName string
	N         int // required amount that must match, -1 for all; so if you have 10 components, and say n=3; 3 must match
	Min       int // the min required amount of structures; so n=-1 and min=0 would pass with no components, as 'all' pass
}

type Stage struct {
	Name string

	PropertyGate  map[string]any // caveat: values must be Go comparable
	ComponentGate []*ComponentGate

	position int
	previous *Stage
	next     *Stage
}

type anyValue struct{ any bool }

// Any represents any value; and should be used in PropertyGate if value is irrelevant, just existence is key
func Any() anyValue {
	return anyValue{true}
}

type Entity interface {
	ID() string
	Kind() string
	Stage() string

	Components(kind string) []Entity // it is required that kind="" returns all components regardless of kind

	HasProperty(name string) bool
	Property(name string) any
}

type Sequence struct {
	current *Stage
	first   *Stage
	lookup  map[string]*Stage
}

func (s *Sequence) Progress() *Sequence {
	if s.current.next != nil {
		return s.FromStage(s.current.next.Name)
	}

	return nil
}

func (s *Sequence) FromStage(name string) *Sequence {
	if name == "" {
		name = s.first.Name
	}

	stage, ok := s.lookup[name]
	if !ok {
		return nil
	}

	return &Sequence{
		lookup:  s.lookup,
		current: stage,
		first:   s.first,
	}
}

func (s *Sequence) Add(next *Stage) *Sequence {
	pos := 1
	s.lookup[next.Name] = next
	if s.current != nil {
		previous := s.current
		pos = previous.position + 1
		next.previous = previous
		previous.next = next
	} else {
		s.first = next
	}
	s.current = next
	next.position = pos
	return s
}

type Orchestration map[string]*Sequence

func (o Orchestration) For(kind string) *Sequence {
	s := &Sequence{
		lookup: map[string]*Stage{},
	}
	o[kind] = s
	return s
}

// Try runs an attempt to move the entity e from current stage to ms, milestone stage
// it'll write to response every pulse that the result is at all gates, including sub-entities transitions
func (o Orchestration) Try(ctx context.Context, e Entity, ms string, response chan Gate) bool {
	kind := e.Kind()
	s, sok := o[kind]
	if !sok {
		select {
		case response <- Gate{
			EntityKind:      e.Kind(),
			EntityID:        e.ID(),
			FromStage:       e.Stage(),
			ToStage:         ms,
			Passes:          false,
			Reason:          fmt.Sprintf("sequence for '%s' not available.", e.Kind()),
			ReasonReference: "sequence_not_available",
		}:
		case <-ctx.Done():
			return false
		}
		return false
	}

	if e.Stage() != "" { // not assigned, will move to start
		_, cok := s.lookup[e.Stage()]
		if !cok {
			select {
			case response <- Gate{
				EntityKind:      e.Kind(),
				EntityID:        e.ID(),
				FromStage:       e.Stage(),
				ToStage:         ms,
				Passes:          false,
				Reason:          fmt.Sprintf("stage '%s' not available for sequence for '%s'.", e.Stage(), e.Kind()),
				ReasonReference: "stage_not_defined",
			}:
			case <-ctx.Done():
				return false
			}
			return false
		}
	}

	m, mok := s.lookup[ms]
	if !mok {
		select {
		case response <- Gate{
			EntityKind:      e.Kind(),
			EntityID:        e.ID(),
			FromStage:       e.Stage(),
			ToStage:         ms,
			Passes:          false,
			Reason:          fmt.Sprintf("milestone '%s' not available for sequence for '%s'.", ms, e.Kind()),
			ReasonReference: "milestone_stage_not_defined",
		}:
		case <-ctx.Done():
			return false
		}
		return false
	}

	// initialize sequence at current entity stage; if empty start with first
	seq := s.FromStage(e.Stage())

	if seq == nil {
		select {
		case response <- Gate{
			EntityKind:      e.Kind(),
			EntityID:        e.ID(),
			FromStage:       e.Stage(),
			ToStage:         ms,
			Passes:          false,
			Reason:          fmt.Sprintf("sequence for '%s' could not be initialized for stage '%s'.", e.Kind(), e.Stage()),
			ReasonReference: "sequence_not_initialized",
		}:
		case <-ctx.Done():
			return false
		}
		return false
	}

	// empty pass success, move from nothing to start
	if e.Stage() == "" && seq.current.Name != "" {
		select {
		case response <- Gate{
			EntityKind: e.Kind(),
			EntityID:   e.ID(),
			FromStage:  "",
			ToStage:    seq.current.Name,
			Passes:     true,
		}:
		case <-ctx.Done():
			return false
		}
	}

	if seq.current.position > m.position {
		select {
		case response <- Gate{
			EntityKind:      e.Kind(),
			EntityID:        e.ID(),
			FromStage:       e.Stage(),
			ToStage:         ms,
			Passes:          false,
			Reason:          fmt.Sprintf("milestone '%s' is in the past of current stage '%s'", ms, e.Stage()),
			ReasonReference: "milestone_stage_in_the_past",
		}:
		case <-ctx.Done():
			return false
		}
		return false
	}

	success := true
	cmp := func(a, b any) bool {
		defer func() { recover() }() // this is just a guard in case the readme was glanced over regarding comparable types
		return a == b
	}

	for seq.current.Name != ms {
		// retrieve next stage and process the progression; if there is no next and milestone wasn't reach exit
		next := seq.Progress()
		succeeded := true

		if next == nil {
			succeeded = false
			success = false
			select {
			case response <- Gate{
				EntityKind:      e.Kind(),
				EntityID:        e.ID(),
				FromStage:       e.Stage(),
				ToStage:         ms,
				Passes:          false,
				Reason:          fmt.Sprintf("milestone '%s' could not be reached.", ms),
				ReasonReference: "milestone_stage_not_reachable",
			}:
			case <-ctx.Done():
				return false
			}
			return false
		}

		for k, v := range next.current.PropertyGate {
			_, isAny := v.(anyValue)
			if !e.HasProperty(k) ||
				(e.HasProperty(k) && !isAny && !cmp(v, e.Property(k))) {
				succeeded = false
				success = false
				select {
				case response <- Gate{
					EntityKind:      e.Kind(),
					EntityID:        e.ID(),
					FromStage:       seq.current.Name,
					ToStage:         next.current.Name,
					Passes:          false,
					Reason:          fmt.Sprintf("transition to stage '%s' requires property '%v' to be set.", next.current.Name, k),
					ReasonReference: fmt.Sprintf("property_gate_not_passed,%s", k),
				}:
				case <-ctx.Done():
					return false
				}
			}
		}

		if len(next.current.ComponentGate) > 0 {
			for _, cg := range next.current.ComponentGate {
				passing := 0
				comps := e.Components(cg.Kind)
				for _, c := range comps {
					componentPasses := o.Try(ctx, c, cg.StageName, response)
					if componentPasses {
						passing++
					}
				}

				n := cg.N
				if (n == -1 && len(comps) > passing) || // if all, but passing is fewer
					(n != -1 && passing != n) || // if not all; and passing doesn't match required
					len(comps) < cg.Min { // if count doesn't suffice required min components
					msg := fmt.Sprintf("%d (min. %d) components in state '%s' required, have %d.", n, cg.Min, cg.StageName, passing)
					if n == -1 {
						msg = fmt.Sprintf("All (min. %d) components in state '%s' required, have %d.", cg.Min, cg.StageName, passing)
					}
					succeeded = false
					success = false
					select {
					case response <- Gate{
						EntityKind:      e.Kind(),
						EntityID:        e.ID(),
						FromStage:       seq.current.Name,
						ToStage:         next.current.Name,
						Passes:          false,
						Reason:          msg,
						ReasonReference: fmt.Sprintf("component_gate_not_passed,%s,%d", cg.StageName, cg.N),
					}:
					case <-ctx.Done():
						return false
					}
				}
			}
		}

		if succeeded {
			select {
			case response <- Gate{
				EntityKind: e.Kind(),
				EntityID:   e.ID(),
				FromStage:  seq.current.Name,
				ToStage:    next.current.Name,
				Passes:     true,
			}:
			case <-ctx.Done():
				return false
			}
		}

		seq = next
	}

	return success
}
