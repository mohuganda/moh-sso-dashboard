package service

import "github.com/moh-sso-dashboard/internal/model"

type Registry struct {
	processors map[model.ProcessType]Processor
}

func NewRegistry() *Registry {
	return &Registry{
		processors: make(map[model.ProcessType]Processor),
	}
}

func (r *Registry) Register(t model.ProcessType, p Processor) {
	r.processors[t] = p
}

func (r *Registry) Get(t model.ProcessType) (Processor, bool) {
	p, ok := r.processors[t]
	return p, ok
}
