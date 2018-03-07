package chains

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

type chainh1 struct {
	called bool
}

func (c *chainh1) HandleRequest(ctx context.Context, wr http.ResponseWriter, rq *http.Request, cx ChainCtx) error {
	c.called = true
	return cx.Next(ctx, wr, rq)
}

type chainerr struct{}

func (c chainerr) HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, cx ChainCtx) error {
	return errors.New("err")
}

func TestChain(t *testing.T) {
	var a, b, c, d chainh1

	chain := Chain(&a).Chain(&b).Chain(&c)
	if err := chain.HandleRequest(context.Background(), nil, nil, chain.elements); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !a.called || !b.called || !c.called || d.called {
		t.Error("Call error")
	}
}

func TestChainErr(t *testing.T) {
	var a, b, c, d chainh1
	var e chainerr

	chain := Chain(&a).Chain(&b).Chain(&e).Chain(&c)
	if err := chain.HandleRequest(context.Background(), nil, nil, chain.elements); err == nil {
		t.Errorf("Expected error")
	}
	if !a.called || !b.called || c.called || d.called {
		t.Error("Call error")
	}
}
