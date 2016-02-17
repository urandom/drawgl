package drawgl

import (
	"encoding/json"
	"sync"

	"github.com/urandom/graph"
)

type OperationJSONConstructor func(opts json.RawMessage) (graph.Linker, error)

var (
	operationsMu sync.Mutex
	operations   = make(map[string]OperationJSONConstructor)
)

func RegisterOperation(name string, constructor OperationJSONConstructor) {
	operationsMu.Lock()
	defer operationsMu.Unlock()

	if constructor == nil {
		panic("drawgl: Register operation constructor is nil")
	}

	if _, dup := operations[name]; dup {
		panic("drawgl: Register called twice for constructor " + name)
	}

	operations[name] = constructor
}
