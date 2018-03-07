This is a simple middleware chain for HTTP handlers:

Each handler can be a function or a struct:
```
func MyHandler(ctx context.Context,w http.ResponseWriter, r *http.Request, cx ChainCtx) error
    ...
    if err:=cx.Next(ctx,w,r); err!=nil {
        return err
    }
    ...
    return nil
}

type AnotherHandler struct {
    ...
}

func (h AnotherHandler) HandleRequest(ctx context.Context,w http.ResponseWriter,r *http.Request, cx ChainCtx) error {
    ...
    if err:=cx.Next(ctx,w,r); err!=nil {
        return err
    }
    ...
    return nil
 }
 ```

You can create a chain using Chain() and ChainFunc() functions:
```
  c:=Chain(StructHandler{...}).ChainFunc(HandlerFunc).Chain(AnotherStruct{})
```
 The chain also contains an error handler:
```
  c=c.Err(func(w http.ResponseWriter,error) {
              // Write error to w
          })
```

