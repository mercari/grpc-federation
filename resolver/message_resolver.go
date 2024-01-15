package resolver

func (g *SequentialVariableDefinitionGroup) VariableDefinitions() VariableDefinitions {
	var defs VariableDefinitions
	if g.Start != nil {
		defs = append(defs, g.Start.VariableDefinitions()...)
	}
	defs = append(defs, g.End)
	return defs
}

func (g *ConcurrentVariableDefinitionGroup) VariableDefinitions() VariableDefinitions {
	var defs VariableDefinitions
	for _, start := range g.Starts {
		defs = append(defs, start.VariableDefinitions()...)
	}
	defs = append(defs, g.End)
	return defs
}
