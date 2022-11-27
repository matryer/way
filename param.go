package way

import (
	"context"
	"sync"
)

type param struct{ key, value string }

type paramPool struct {
	pool *sync.Pool
}

func (p paramPool) Get() *[]param {
	params := p.pool.Get().(*[]param)
	*params = (*params)[:0]
	return params
}

func (p paramPool) Put(params *[]param) {
	if params == nil {
		return
	}
	p.pool.Put(params)
}

var parpool = paramPool{
	pool: &sync.Pool{New: func() interface{} {
		p := make([]param, 12)
		return &p
	}},
}

type paramkey struct{}

func toParamContext(ctx context.Context, params *[]param) context.Context {
	return context.WithValue(ctx, paramkey{}, params)
}

// Param gets the path parameter from the specified Context.
// Returns an empty string if the parameter was not found.
func Param(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	params, _ := ctx.Value(paramkey{}).(*[]param)
	if params == nil {
		return ""
	}

	for _, param := range *params {
		if param.key == key {
			return param.value
		}
	}
	return ""
}
