package resolver

// refreshServiceCELPlugins re-populates Service.CELPlugins from the owning
// file's AllCELPlugins after markOptionImportReachable runs. resolveService
// captures CELPlugins at service-creation time, which happens during
// resolveFile — before the per-compile reachability closure exists. Without
// this refresh, every Service.CELPlugins would be empty whenever the gate
// would otherwise have included a plugin, since AllCELPlugins returns no
// option-import-reachable plugins during that earlier pass.
func (r *Resolver) refreshServiceCELPlugins(files []*File) {
	for _, f := range files {
		for _, svc := range f.Services {
			svc.CELPlugins = f.AllCELPlugins()
		}
	}
}

// markOptionImportReachable computes the per-compile option-import
// reachability of each resolved file and stamps the IsImportedByOption flag
// accordingly. A file is option-import-reachable iff some file in this
// compile graph reaches it through an (grpc.federation.file).import edge —
// directly or transitively via other option-import edges. AllCELPlugins
// uses the flag to gate plugin auto-registration: a file's plugin.export
// contributes to every CEL env in this compile iff the file is
// option-import-reachable.
func (r *Resolver) markOptionImportReachable(files []*File) {
	reachable := make(map[*File]bool)
	var visit func(f *File)
	visit = func(f *File) {
		if reachable[f] {
			return
		}
		reachable[f] = true
		for _, dep := range f.OptionImports {
			visit(dep)
		}
	}
	for _, f := range files {
		for _, dep := range f.OptionImports {
			visit(dep)
		}
	}
	for f := range reachable {
		f.IsImportedByOption = true
	}
}
