package pipeline

import (
	"context"
	"github.com/topfreegames/pitaya/agent"
	"github.com/topfreegames/pitaya/conn/message"
)

var (
	// BeforeFilterHandler contains the functions to be called before the handler method is executed
	BeforeFilterHandler = &FilterChannel{}
	// AfterFilterHandler contains the functions to be called after the handler method is executed
	AfterFilterHandler = &FilterAfterChannel{}
)

type (
	FilterChannel struct {
		Filters []FilterHandler
	}

	FilterAfterChannel struct {
		Filters []FilterAfterHandler
	}

	FilterHandler      func(ctx context.Context, agent *agent.Agent, message *message.Message) error
	FilterAfterHandler func(ctx context.Context, agent *agent.Agent, message *message.Message) error
)

// PushBack should not be used after pitaya is running
func (p *FilterChannel) PushBack(h FilterHandler) {
	p.Filters = append(p.Filters, h)
}

// PushBack should not be used after pitaya is running
func (p *FilterAfterChannel) PushBack(h FilterAfterHandler) {
	p.Filters = append(p.Filters, h)
}
