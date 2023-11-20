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
		for _, resolver := range method.Response.CustomResolvers() {
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
		for _, svc := range method.Response.DependServices() {
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
