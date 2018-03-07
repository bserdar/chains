// Package chains implements a simple HTTP server middleware chain.
//
// Each handler can be a function or a struct:
//      func MyHandler(ctx context.Context,w http.ResponseWriter, r *http.Request, cx ChainCtx) error {
//        ...
//        if err:=cx.Next(ctx,w,r); err!=nil {
//            return err
//        }
//        ...
//        return nil
//      }
//
//
//    type AnotherHandler struct {
//        ...
//    }
//
//    func (h AnotherHandler) HandleRequest(ctx context.Context,w http.ResponseWriter,r *http.Request, cx ChainCtx) error {
//        ...
//        if err:=cx.Next(ctx,w,r); err!=nil {
//            return err
//        }
//        ...
//        return nil
//    }
//
// You can create a chain using Chain() and ChainFunc() functions:
//    c:=Chain(StructHandler{...}).ChainFunc(HandlerFunc).Chain(AnotherStruct{})
// The chain also contains an error handler:
//    c=c.Err(func(w http.ResponseWriter,error) {
//                // Write error to w
//            })
//
//
//
package chains

import (
	"context"
	"net/http"
)

// ChainCtx is passed to chain handler functions
type ChainCtx []Handler

// Next calls the next handler in chain
func (c ChainCtx) Next(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if len(c) > 0 {
		return c[0].HandleRequest(ctx, w, r, c[1:])
	}
	return nil
}

// HandlerFunc defines the structure of a handler function
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request, ChainCtx) error

// RenderErrorFunc renders error response
type RenderErrorFunc func(http.ResponseWriter, error)

// Handler handles a request, and may return an error
type Handler interface {
	HandleRequest(context.Context, http.ResponseWriter, *http.Request, ChainCtx) error
}

// HandlerChain performs the basic chain functionality
type HandlerChain struct {
	renderErrorFunc RenderErrorFunc
	elements        []Handler
}

type handlerFunc struct {
	f HandlerFunc
}

// HandleRequest calls the handle func
func (h handlerFunc) HandleRequest(ctx context.Context, writer http.ResponseWriter, request *http.Request, cx ChainCtx) error {
	return h.f(ctx, writer, request, cx)
}

// MakeHandler returns a handler from a function
func MakeHandler(f HandlerFunc) Handler {
	return handlerFunc{f: f}
}

// BasicErrorHandler sets the http header, doesn't return body
func BasicErrorHandler(writer http.ResponseWriter, err error) {
	writer.WriteHeader(http.StatusInternalServerError)
}

// Chain creates a new chain and adds the handler to the chain
func Chain(h Handler) *HandlerChain {
	chain := HandlerChain{renderErrorFunc: BasicErrorHandler,
		elements: make([]Handler, 0)}
	chain.Chain(h)
	return &chain
}

// ChainFunc creates a new chain and adds the function to the chain
func ChainFunc(f HandlerFunc) *HandlerChain {
	return Chain(MakeHandler(f))
}

// Err sets the error render function
func (chain *HandlerChain) Err(f RenderErrorFunc) *HandlerChain {
	chain.renderErrorFunc = f
	return chain
}

// Chain adds a new handler to the chain
func (chain *HandlerChain) Chain(h Handler) *HandlerChain {
	chain.elements = append(chain.elements, h)
	return chain
}

// ChainFunc adds a function to the chain
func (chain *HandlerChain) ChainFunc(f HandlerFunc) *HandlerChain {
	return chain.Chain(MakeHandler(f))
}

// HandleRequest calls all the elements of the chain until one fails, or all are done
func (chain *HandlerChain) HandleRequest(ctx context.Context, writer http.ResponseWriter, request *http.Request, cx ChainCtx) error {
	return cx.Next(ctx, writer, request)
}

// ServeHTTP of http.Handler
func (chain *HandlerChain) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	err := chain.HandleRequest(request.Context(), writer, request, chain.elements)
	if err != nil {
		chain.renderErrorFunc(writer, err)
	}
}
