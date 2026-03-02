package service

type Registry struct {
	processors map[string]Processor
}

func NewRegistry() *Registry {
	return &Registry{
		processors: make(map[string]Processor),
	}
}

func (r *Registry) Register(processType string, p Processor) {
	r.processors[processType] = p
}

func (r *Registry) Get(processType string) (Processor, bool) {
	p, ok := r.processors[processType]
	return p, ok
}
