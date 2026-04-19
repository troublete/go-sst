package sst

const (
	RequiredPropertyMissing = "required_property_missing"
	RequiredPropertyNoMatch = "required_property_no_match"
)

type PropertyGateDefinition struct {
	keys, values []string

	logicGate string // this is for internal use, to reference logic gates when combining inputs; todo: potentially change to slice
}

func NewPropertyGate(in map[string]string, logicGate ...string) *PropertyGateDefinition {
	// as the values are checked sequentially we change the format here to slices
	var keys []string
	var values []string
	for k, v := range in {
		keys = append(keys, k)
		values = append(values, v)
	}
	var lg string
	if len(logicGate) > 0 {
		lg = logicGate[0]
	}
	return &PropertyGateDefinition{
		keys:      keys,
		values:    values,
		logicGate: lg,
	}
}

func (pgd *PropertyGateDefinition) Evaluate(on map[string]string, postfix string) (bool, []string) {
	var issues []string

	// we serialized keys to slice on creation, so now we use this to use maps strength, lookup to check for values
	for idx, k := range pgd.keys {
		// if either the value is not set; or the value is set and the required is not empty (indicating any value) but does
		// not match, we fail
		if onV, ok := on[k]; !ok || (onV != pgd.values[idx] && pgd.values[idx] != "") {
			if !ok {
				issues = append(issues, RequiredPropertyMissing+",key="+k+","+postfix)
			} else {
				issues = append(issues, RequiredPropertyNoMatch+",key="+k+","+postfix)
			}
		}
	}

	b := len(issues) == 0
	return b, issues
}
