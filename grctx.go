package grctx

import (
	"fmt"
	"hash/crc64"
	"math/rand"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type context struct {
	ctx interface{}
}

type globalContext struct {
	ctxs map[uint64]*context
	lock *sync.RWMutex
}

var commonFuncRe = regexp.MustCompile(`\bgrctx\.WithContext\.func1\b`)
var crcTable = crc64.MakeTable(crc64.ECMA)
var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

var gCtx = globalContext{
	ctxs: make(map[uint64]*context),
	lock: &sync.RWMutex{},
}

func getStackFrame() string {
	lines := strings.Split(string(debug.Stack()), "\n")
	for _, l := range lines {
		if commonFuncRe.MatchString(l) {
			return l
		}
	}

	return ""
}

func getKey() uint64 {
	frameStr := getStackFrame()
	if frameStr == "" {
		return 0
	}

	return crc64.Checksum([]byte(frameStr), crcTable)
}

func (gc *globalContext) store(key uint64, ctx *context) {
	gc.lock.Lock()
	defer gc.lock.Unlock()

	if _, ok := gc.ctxs[key]; ok {
		// XXX: remove panic
		panic("Duplicated key")
	}

	gc.ctxs[key] = ctx
}

func (gc *globalContext) read(key uint64) (*context, bool) {
	gc.lock.RLock()
	defer gc.lock.RUnlock()

	ctx, ok := gc.ctxs[key]
	return ctx, ok
}

func (gc *globalContext) destroy(key uint64) {
	gc.lock.Lock()
	defer gc.lock.Unlock()

	if _, ok := gc.ctxs[key]; !ok {
		// XXX: remove panic
		panic("Key wasn't found in storage")
	}

	delete(gc.ctxs, key)
}

func WithContext(fn func(), ctx interface{}) {
	randN := uint64(randSource.Int63())
	ctxWrp := context{
		ctx: ctx,
	}

	// Following function will be used to calculate unique hash sum that will be used to identify
	// the global context for function fn. randN variable is used just to be sure the sum calculeated
	// using stack frame information is really unique.
	func(fn func(), ctx *context, _ uint64) {
		key := getKey()
		if key != 0 {
			gCtx.store(key, ctx)
			defer gCtx.destroy(key)
		}

		fn()
	}(fn, &ctxWrp, randN)
}

func Context() (interface{}, error) {
	key := getKey()
	if key == 0 {
		return nil, fmt.Errorf("context not found (possibly function wasn't called by WithContext func)")
	}

	data, ok := gCtx.read(key)
	if !ok {
		return nil, fmt.Errorf("context was already destroyed")
	}

	return data.ctx, nil
}
