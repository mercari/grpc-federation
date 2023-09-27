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

func (s *Service) CustomResolvers() []*CustomResolver {
	var resolvers []*CustomResolver
	for _, method := range s.Methods {
		resolvers = append(resolvers, s.customResolversByMessage(method.Response)...)
	}
	sort.Slice(resolvers, func(i, j int) bool {
		return resolvers[i].Message.Name < resolvers[j].Message.Name
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
	serviceMap := map[*Service]struct{}{}
	for _, method := range s.Methods {
		rule := method.Response.Rule
		if rule == nil {
			continue
		}
		for _, group := range rule.Resolvers {
			for _, resolver := range group.Resolvers() {
				for _, svc := range s.useServices(resolver) {
					serviceMap[svc] = struct{}{}
				}
			}
		}
	}
	services := make([]*Service, 0, len(serviceMap))
	for service := range serviceMap {
		services = append(services, service)
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})
	return services
}

func (s *Service) useServices(resolver *MessageResolver) []*Service {
	var services []*Service
	if resolver.MethodCall != nil {
		methodCall := resolver.MethodCall
		services = append(services, methodCall.Method.Service)
	}
	msgResolverMap := make(map[*MessageResolver]struct{})
	if resolver.MessageDependency != nil && resolver.MessageDependency.Message != nil && resolver.MessageDependency.Message.Rule != nil {
		for _, group := range resolver.MessageDependency.Message.Rule.Resolvers {
			for _, resolver := range group.Resolvers() {
				services = append(services, s.useServicesRecursive(resolver, msgResolverMap)...)
			}
		}
	}
	return services
}

func (s *Service) useServicesRecursive(resolver *MessageResolver, msgResolverMap map[*MessageResolver]struct{}) []*Service {
	if _, found := msgResolverMap[resolver]; found {
		return nil
	}
	msgResolverMap[resolver] = struct{}{}

	var services []*Service
	if resolver.MethodCall != nil {
		methodCall := resolver.MethodCall
		services = append(services, methodCall.Method.Service)
	}
	if resolver.MessageDependency != nil && resolver.MessageDependency.Message != nil && resolver.MessageDependency.Message.Rule != nil {
		for _, group := range resolver.MessageDependency.Message.Rule.Resolvers {
			for _, resolver := range group.Resolvers() {
				services = append(services, s.useServicesRecursive(resolver, msgResolverMap)...)
			}
		}
	}
	return services
}
