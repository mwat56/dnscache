/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package nodepool

import (
	"errors"
	"runtime"
	"sync/atomic"
	"time"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// `poolDropMask` is the bit mask to use for limiting the returns
	// to the pool.
	poolDropMask = 7 // 111

	// `poolInitSize` is the number of nodes to pre-allocate
	// for the pool during initialisation.
	//
	// Four times this value is used as the pool's maximum size.
	poolInitSize = 1 << 9 // 512
)

type (
	// `TPoolMetrics` contains the metrics data for the pool.
	//
	// These are the fields providing the metrics data:
	//
	//   - `Created`: Number of nodes created by the pool.
	//   - `Returned`: Number of nodes returned to the pool.
	//   - `Size`: Current number of nodes in the pool.
	TPoolMetrics struct {
		Created  uint32
		Returned uint32
		Size     int
	}

	// `TPool` is a bounded pool of nodes.
	//
	// The pool is inherently thread-safe. Its size is fixed and can't
	// be changed after creation.
	//
	// The pool's `New()` method is called to create new nodes.
	//
	TPool struct {
		New      func() any        // Factory function for nodes
		nodes    chan any          // Bounded channel for nodes
		cCh      chan uint32       // Channel for newly created nodes
		rCh      chan uint32       // Channel for returned nodes
		sCh      chan int          // Channel for pool size
		mCh      chan TPoolMetrics // Channel for pool metrics
		created  atomic.Uint32     // Number of nodes created
		returned atomic.Uint32     // Number of nodes returned
	}

	// `TPoolError` is returned if the pool is not fully initialised.
	TPoolError struct {
		error
	}
)

var (
	// `ErrNilFactory` is returned if the factory function is nil.
	ErrNilFactory = TPoolError{errors.New("nil factory function `New()` in Pool")}

	// `ErrPoolNotInit` is returned if the pool is not fully initialised.
	ErrPoolNotInit = TPoolError{errors.New("node pool not initialised")}
)

// ---------------------------------------------------------------------------
// Pool initialisation:

// `Init()` initialises the pool.
//
// Note: This function should be called before any other pool-related actions
// to ensure proper initialisation.
//
// The created pool is inherently thread-safe.
//
// The required `aNewFunc` function is called to create new nodes if the
// pool runs out of cached nodes.
//
// The pool's size is fixed and can't be changed after initialisation.
//
// Parameters:
//   - `aNewFunc`: Factory function for creating new nodes.
//   - `aSize`: Initial size of the pool.
func Init(aNewFunc func() any, aSize int) (rPool *TPool, rErr error) {
	if nil == aNewFunc {
		rErr = ErrNilFactory
		return
	}
	if 0 >= aSize {
		aSize = poolInitSize
	}

	rPool = &TPool{New: aNewFunc}
	rPool.reset(aSize)

	return
} // Init()

// ---------------------------------------------------------------------------
// `TPool` methods:

// `CreatedChannel()` returns a R/O channel that provides the current number
// of nodes created by the pool.
//
// The channel can be used for real-time monitoring of the pool's work.
//
// Returns:
//   - `rChan`: Channel that provides the number of nodes created.
//   - `rErr`: An error, if any.
func (p *TPool) CreatedChannel() (rChan <-chan uint32, rErr error) {
	if nil == p {
		rErr = ErrPoolNotInit
		return
	}
	if nil == p.New {
		rErr = ErrNilFactory
		return
	}
	if nil == p.cCh {
		// Pool not initialised yet
		p.reset(0)
	}
	rChan = p.cCh

	return
} // CreatedChannel()

// `Get()` picks a node from the pool.
//
// Calling this method is the "raw" version of getting a node from
// the pool. An improved way would be to create a wrapper function that
// calls this method and then assures that the returned node's fields
// are properly initialised/reset before returning it to the caller.
//
// If the pool is empty, a new node instance is created.
//
// Returns:
//   - `rNode`: A node from the pool.
//   - `rErr`: An error, if any.
func (p *TPool) Get() (rNode any, rErr error) {
	if nil == p {
		rErr = ErrPoolNotInit
		return
	}
	if nil == p.New {
		rErr = ErrNilFactory
		return
	}
	if nil == p.nodes {
		// Pool not initialised yet
		p.reset(0)
	}

	select {
	case rNode = <-p.nodes:
		// Node was taken from pool
	default:
		rNode = p.newNode()
	}

	select {
	case p.sCh <- len(p.nodes):
		// New pool size was written
	default:
		// Ignore if nobody's listening
	}
	go sendMetrics(p, 0, 0)

	return
} // Get()

// `Metrics()` returns the current pool metrics.
//
// Returns:
//   - `Created`: Number of nodes created by the pool.
//   - `Returned`: Number of nodes returned to the pool.
//   - `Size`: Current number of nodes in the pool.
func (p *TPool) Metrics() (rMetric *TPoolMetrics, rErr error) {
	if nil == p {
		rErr = ErrPoolNotInit
		return
	}
	if nil == p.New {
		rErr = ErrNilFactory
		return
	}
	if nil == p.nodes {
		p.reset(0)
	}
	rMetric = &TPoolMetrics{
		Created:  p.created.Load(),
		Returned: p.returned.Load(),
		Size:     len(p.nodes),
	}

	return
} // Metrics()

// `newNode()` creates a new node.
//
// This private method is called internally by the `Get()` method if the pool
// is empty. It utilises the pool's `New()` function to create the new node.
//
// Returns:
//   - `rNode`: A new node.
func (p *TPool) newNode() (rNode any) {
	rNode = p.New()
	//TODO: Go 1.24:
	// runtime.AddCleanup(rNode, func() {
	// 	p.put(rNode)
	// })

	c := p.created.Add(1)
	select {
	case p.cCh <- c:
		// Counter was written
	default:
		// Ignore if nobody's listening
	}

	return
} // newNode()

// `Put()` throws a node back into the pool.
//
// If the pool is full, the node is dropped.
//
// Parameters:
//   - `aNode`: The node to return to the pool.
//
// Returns:
//   - `rErr`: An error, if any.
func (p *TPool) Put(aNode any) (rErr error) {
	if nil == p {
		rErr = ErrPoolNotInit
		return
	}
	if nil == p.New {
		rErr = ErrNilFactory
		return
	}
	if nil == p.nodes {
		// Pool not initialised yet
		p.reset(0)
	}

	r := p.returned.Add(1)
	select {
	case p.rCh <- r:
		// Counter was written
	default:
		// Ignore if nobody's listening
	}

	if (r & poolDropMask) == poolDropMask {
		// Drop the node if the drop mask matches.
		// This leaves the given `aNode` for GC.
		// With a drop mask of `7` (111) we drop 1 in 8 nodes.
		return
	}

	select {
	case p.nodes <- aNode:
		// Node was returned to the pool
	default:
		// Drop if pool is full
	}

	select {
	case p.sCh <- len(p.nodes):
		// New pool size was written
	default:
		// Ignore if nobody's listening
	}
	go sendMetrics(p, 0, r)

	return
} // Put()

// `reset()` resets the pool to its initial state.
//
// Parameters:
//   - `aSize`: Initial size of the pool.
func (p *TPool) reset(aSize int) {
	if 0 >= aSize {
		aSize = poolInitSize
	}
	p.nodes = make(chan any, aSize<<2)
	p.cCh = make(chan uint32, 1)       // new nodes created
	p.rCh = make(chan uint32, 1)       // returned nodes
	p.sCh = make(chan int, 1)          // pool size
	p.mCh = make(chan TPoolMetrics, 1) // pool metrics

	// Pre-allocate some nodes for the pool:
	for range aSize {
		p.Put(p.New())
	}
} // reset()

// `ReturnedChannel()` returns a R/O channel that provides the current number
// of nodes returned to the pool.
//
// The channel can be used for real-time monitoring of the pool's work.
//
// Returns:
//   - `rChan`: Channel that provides the number of nodes returned.
//   - `rErr`: An error, if any.
func (p *TPool) ReturnedChannel() (rChan <-chan uint32, rErr error) {
	if nil == p {
		rErr = ErrPoolNotInit
		return
	}
	if nil == p.New {
		rErr = ErrNilFactory
		return
	}
	if nil == p.rCh {
		// Pool not initialised yet
		p.reset(0)
	}
	rChan = p.rCh

	return
} // ReturnedChannel()

// `sendMetrics()` sends the pool's metrics to the metrics channel.
//
// This function is called asynchronously by the `Get()` and `Put()` methods.
//
// Parameters:
//   - `aPool`: The pool to send the metrics for.
//   - `aCreate`: Number of nodes created (if 0, use pool's counter).
//   - `aReturn`: Number of nodes returned (if 0, use pool's counter).
func sendMetrics(aPool *TPool, aCreate, aReturn uint32) {
	if 0 == aCreate {
		aCreate = aPool.created.Load()
	}
	if 0 == aReturn {
		aReturn = aPool.returned.Load()
	}
	m := TPoolMetrics{
		Created:  aCreate,
		Returned: aReturn,
		Size:     len(aPool.nodes),
	}

	select {
	case aPool.mCh <- m:
		// Counters were written
		runtime.Gosched()
	case <-time.After(time.Second):
		// Timeout (just to be sure to not block)
		return
	default:
		// Ignore if nobody's listening
	}
} // sendMetrics()

// `SizeChannel()` returns a R/O channel that provides the current number
// of nodes in the pool.
//
// The channel can be used for real-time monitoring of the pool's work.
//
// Returns:
//   - `rChan`: Channel that provides the number of nodes in the pool.
//   - `rErr`: An error, if any.
func (p *TPool) SizeChannel() (rChan <-chan int, rErr error) {
	if nil == p {
		rErr = ErrPoolNotInit
		return
	}
	if nil == p.New {
		rErr = ErrNilFactory
		return
	}
	if nil == p.sCh {
		// Pool not initialised yet
		p.reset(0)
	}
	rChan = p.sCh

	return
} // SizeChannel()

/* _EoF_ */
