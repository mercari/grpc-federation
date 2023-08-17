package resolver

func (g *SequentialMessageResolverGroup) Resolvers() []*MessageResolver {
	var resolvers []*MessageResolver
	if g.Start != nil {
		resolvers = append(resolvers, g.Start.Resolvers()...)
	}
	resolvers = append(resolvers, g.End)
	return resolvers
}

func (g *ConcurrentMessageResolverGroup) Resolvers() []*MessageResolver {
	var resolvers []*MessageResolver
	for _, start := range g.Starts {
		resolvers = append(resolvers, start.Resolvers()...)
	}
	resolvers = append(resolvers, g.End)
	return resolvers
}
