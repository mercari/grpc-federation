package resolver

import (
	"fmt"
	"sort"
)

func (s *Service) GoPackage() *GoPackage {
	if s.File == nil {
		return nil
	}
	return s.File.GoPackage
}

func (s *Service) Package() *Package {
	if s.File == nil {
		return nil
	}
	return s.File.Package
}

func (s *Service) FQDN() string {
	return fmt.Sprintf("%s.%s", s.PackageName(), s.Name)
}

func (s *Service) PackageName() string {
	pkg := s.Package()
	if pkg == nil {
		return ""
	}
	return pkg.Name
}

func (s *Service) Method(name string) *Method {
	for _, method := range s.Methods {
		if method.Name == name {
			return method
		}
	}
	return nil
}

func (s *Service) HasMessageInMethod(msg *Message) bool {
	for _, mtd := range s.Methods {
		if mtd.Request == msg {
			return true
		}
		if mtd.Response == msg {
			return true
		}
	}
	return false
}

func (s *Service) GoPackageDependencies() []*GoPackage {
	pkgMap := map[*GoPackage]struct{}{}
	pkgMap[s.GoPackage()] = struct{}{}
	for _, dep := range s.ServiceDependencies() {
		pkgMap[dep.Service.GoPackage()] = struct{}{}
	}
	pkgs := make([]*GoPackage, 0, len(pkgMap))
	for pkg := range pkgMap {
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

type CustomResolver struct {
	Message *Message
	Field   *Field
}

func (r *CustomResolver) FQDN() string {
	if r.Field != nil {
		return fmt.Sprintf("%s.%s", r.Message.FQDN(), r.Field.Name)
	}
	return r.Message.FQDN()
}

func (s *Service) CustomResolvers() []*CustomResolver {
	resolverMap := make(map[string]*CustomResolver)
	for _, method := range s.Methods {
		for _, resolver := range s.customResolversByMessage(method.Response) {
			resolverMap[resolver.FQDN()] = resolver
		}
	}
	resolvers := make([]*CustomResolver, 0, len(resolverMap))
	for _, resolver := range resolverMap {
		resolvers = append(resolvers, resolver)
	}
	sort.Slice(resolvers, func(i, j int) bool {
		return resolvers[i].FQDN() < resolvers[j].FQDN()
	})
	return resolvers
}

func (s *Service) ExistsCustomResolver() bool {
	return len(s.CustomResolvers()) != 0
}

func (s *Service) customResolversByMessage(msg *Message) []*CustomResolver {
	var crs []*CustomResolver
	if msg.HasCustomResolver() {
		crs = append(crs, &CustomResolver{Message: msg})
	}
	for _, field := range msg.Fields {
		if field.HasCustomResolver() {
			crs = append(crs, &CustomResolver{
				Message: msg,
				Field:   field,
			})
		}
	}
	for _, group := range msg.Rule.Resolvers {
		for _, resolver := range group.Resolvers() {
			crs = append(crs, s.customResolvers(resolver)...)
		}
	}
	return crs
}

func (s *Service) customResolvers(resolver *MessageResolver) []*CustomResolver {
	var customResolvers []*CustomResolver
	dep := resolver.MessageDependency
	if dep != nil {
		customResolvers = append(customResolvers, s.customResolversByMessage(dep.Message)...)
	}
	return customResolvers
}

func (s *Service) ServiceDependencies() []*ServiceDependency {
	useServices := s.UseServices()
	deps := make([]*ServiceDependency, 0, len(s.Rule.Dependencies)+len(useServices))
	depSvcMap := map[string]*ServiceDependency{}
	for _, dep := range s.Rule.Dependencies {
		depSvcMap[dep.Service.FQDN()] = dep
		deps = append(deps, dep)
	}
	for _, svc := range useServices {
		if _, exists := depSvcMap[svc.FQDN()]; !exists {
			deps = append(deps, &ServiceDependency{Service: svc})
		}
	}
	return deps
}

func (s *Service) UseServices() []*Service {
	svcMap := map[*Service]struct{}{}
	for _, method := range s.Methods {
		for _, svc := range s.useServicesByMessage(method.Response, make(map[*MessageResolver]struct{})) {
			svcMap[svc] = struct{}{}
		}
	}
	svcs := make([]*Service, 0, len(svcMap))
	for svc := range svcMap {
		svcs = append(svcs, svc)
	}
	sort.Slice(svcs, func(i, j int) bool {
		return svcs[i].Name < svcs[j].Name
	})
	return svcs
}

func (s *Service) useServicesByMessage(msg *Message, msgResolverMap map[*MessageResolver]struct{}) []*Service {
	if msg == nil {
		return nil
	}

	var svcs []*Service
	if msg.Rule != nil {
		for _, group := range msg.Rule.Resolvers {
			for _, resolver := range group.Resolvers() {
				svcs = append(svcs, s.useServices(resolver, msgResolverMap)...)
			}
		}
	}
	for _, field := range msg.Fields {
		if field.Rule == nil {
			continue
		}
		if field.Rule.Oneof == nil {
			continue
		}
		for _, group := range field.Rule.Oneof.Resolvers {
			for _, resolver := range group.Resolvers() {
				svcs = append(svcs, s.useServices(resolver, msgResolverMap)...)
			}
		}
	}
	return svcs
}

func (s *Service) useServices(resolver *MessageResolver, msgResolverMap map[*MessageResolver]struct{}) []*Service {
	if _, found := msgResolverMap[resolver]; found {
		return nil
	}
	msgResolverMap[resolver] = struct{}{}

	var svcs []*Service
	if resolver.MethodCall != nil {
		methodCall := resolver.MethodCall
		svcs = append(svcs, methodCall.Method.Service)
	}
	if resolver.MessageDependency == nil {
		return svcs
	}
	svcs = append(svcs, s.useServicesByMessage(resolver.MessageDependency.Message, msgResolverMap)...)
	return svcs
}
