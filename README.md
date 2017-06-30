# grctx
Global per-request context

# Usage example

```go
package main

import (
	"fmt"

	"grctx"
)

func test() {
	ctx, err := grctx.Context()
	if err != nil {
		panic(err)
	}

	if data, ok := ctx.(string); ok {
		fmt.Printf("Context is: %q\n", data)
		return
	}

	panic(fmt.Errorf("invalid context: %+v", ctx))
}

func main() {
	grctx.WithContext(test, "Context 1")
	grctx.WithContext(test, "Context 2")
	grctx.WithContext(test, "Context 3")
}
```
