package resolver

import (
	"fmt"
	"sort"
)

func (s *Service) GoPackage() *GoPackage {
	if s == nil {
		return nil
	}
	if s.File == nil {
		return nil
	}
	return s.File.GoPackage
}

func (s *Service) Package() *Package {
	if s == nil {
		return nil
	}
	if s.File == nil {
		return nil
	}
	return s.File.Package
}

func (s *Service) PackageName() string {
	if s == nil {
		return ""
	}
	pkg := s.Package()
	if pkg == nil {
		return ""
	}
	return pkg.Name
}

func (s *Service) Method(name string) *Method {
	if s == nil {
		return nil
	}
	for _, method := range s.Methods {
		if method.Name == name {
			return method
		}
	}
	return nil
}

func (s *Service) HasMessageInMethod(msg *Message) bool {
	if s == nil {
		return false
	}
	for _, mtd := range s.Methods {
		if mtd.Request == msg {
			return true
		}
		if mtd.FederationResponse() == msg {
			return true
		}
	}
	return false
}

func (s *Service) HasMessageInVariables(msg *Message) bool {
	if s.Rule == nil || len(s.Rule.Vars) == 0 {
		return false
	}
	for _, svcVar := range s.Rule.Vars {
		for _, msgExpr := range svcVar.MessageExprs() {
			if msgExpr.Message == msg {
				return true
			}
		}
	}
	return false
}

func (s *Service) GoPackageDependencies() []*GoPackage {
	if s == nil {
		return nil
	}
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
	if r == nil {
		return ""
	}
	if r.Field != nil {
		return fmt.Sprintf("%s.%s", r.Message.FQDN(), r.Field.Name)
	}
	return r.Message.FQDN()
}

func (s *Service) CustomResolvers() []*CustomResolver {
	if s == nil {
		return nil
	}
	resolverMap := make(map[string]*CustomResolver)
	for _, method := range s.Methods {
		for _, resolver := range method.FederationResponse().CustomResolvers() {
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
	if s == nil {
		return false
	}
	return len(s.CustomResolvers()) != 0
}

func (s *Service) ServiceDependencies() []*ServiceDependency {
	if s == nil {
		return nil
	}
	useServices := s.UseServices()
	deps := make([]*ServiceDependency, 0, len(useServices))
	depSvcMap := map[string]*ServiceDependency{}
	for _, svc := range useServices {
		if _, exists := depSvcMap[svc.FQDN()]; !exists {
			deps = append(deps, &ServiceDependency{Service: svc})
		}
	}
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Service.FQDN() < deps[j].Service.FQDN()
	})
	return deps
}

func (s *Service) UseServices() []*Service {
	if s == nil {
		return nil
	}
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
		return svcs[i].FQDN() < svcs[j].FQDN()
	})
	return svcs
}

func (sv *ServiceVariable) MessageExprs() []*MessageExpr {
	if sv.Expr == nil {
		return nil
	}
	expr := sv.Expr
	switch {
	case expr.Map != nil:
		return expr.Map.MessageExprs()
	case expr.Message != nil:
		return []*MessageExpr{expr.Message}
	}
	return nil
}

func (sv *ServiceVariable) ToVariableDefinition() *VariableDefinition {
	def := &VariableDefinition{
		Name: sv.Name,
		If:   sv.If,
	}
	if sv.Expr == nil {
		return def
	}
	expr := sv.Expr
	var validationExpr *ValidationExpr
	if expr.Validation != nil {
		validationExpr = &ValidationExpr{
			Error: &GRPCError{
				If:      expr.Validation.If,
				Message: expr.Validation.Message,
			},
		}
	}
	def.Expr = &VariableExpr{
		Type:       expr.Type,
		By:         expr.By,
		Map:        expr.Map,
		Message:    expr.Message,
		Enum:       expr.Enum,
		Validation: validationExpr,
	}
	return def
}
