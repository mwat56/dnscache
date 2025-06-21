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
	//   - `Size`: Current number of nodes in the pool.
	//   - `Created`: Number of nodes created by the pool.
	//   - `Returned`: Number of nodes returned to the pool.
	TPoolMetrics struct {
		Size     int
		Created  uint32
		Returned uint32
	}

	// `TPool` is a bounded pool of nodes.
	//
	// The pool is inherently thread-safe. Its size is fixed and can't
	// be changed after creation.
	//
	// The optional `New()` method is called to create new nodes, if the pool
	// runs out of cached nodes.
	//
	TPool struct {
		New      func() any         // Factory function for nodes
		nodes    chan any           // Bounded channel for nodes
		mCh      chan *TPoolMetrics // Channel for pool metrics
		created  atomic.Uint32      // Number of nodes created
		returned atomic.Uint32      // Number of nodes returned
	}

	// `TPoolError` is returned if the pool is not fully initialised.
	TPoolError struct {
		error
	}
)

var (
	// `ErrPoolNotInit` is returned if the pool is not fully initialised.
	ErrPoolNotInit = TPoolError{errors.New("node pool not initialised")}
)

// ---------------------------------------------------------------------------
// Pool initialisation:

// `Init()` initialises a new node pool.
//
// Note: This function should be called before any other pool-related actions
// to ensure proper initialisation.
//
// The created pool is inherently thread-safe.
//
// The optional `aNewFunc` function is called to create new nodes if the pool
// runs out of cached nodes. If that function is not provided (i.e. `nil`),
// the pool will not create any new nodes and will return `nil` instead; in
// that case the only way to get nodes into the pool is to use the `Put()`
// method.
//
// The pool's size is fixed and can't be changed after initialisation.
//
// Parameters:
//   - `aNewFunc`: Factory function for creating new nodes.
//   - `aSize`: Initial size of the pool.
func Init(aNewFunc func() any, aSize int) (rPool *TPool, rErr error) {
	if 0 >= aSize {
		aSize = poolInitSize
	}

	rPool = &TPool{}
	if nil != aNewFunc {
		rPool.New = aNewFunc
	}
	rPool.reset(aSize)

	return
} // Init()

// ---------------------------------------------------------------------------
// `TPool` methods:

// `Get()` picks a node from the pool.
//
// Calling this method is the "raw" version of getting a node from the pool.
// An improved way would be to create a wrapper function that calls this
// method and then assures that the returned node's fields are properly
// initialised/reset before passing it to the caller.
//
// If the pool is empty, a new node instance is created if the pool's
// `New()` function is set. Otherwise `nil` is returned.
//
// Returns:
//   - `rNode`: A node from the pool.
//   - `rErr`: An error, if any.
func (p *TPool) Get() (rNode any, rErr error) {
	if nil == p {
		rErr = ErrPoolNotInit
		return
	}
	if nil == p.nodes {
		// Pool not initialised yet
		p.reset(0)
	}
	var c uint32

	select {
	case rNode = <-p.nodes:
		// Node was taken from pool
	default:
		rNode, c = p.newNode()
	}
	go sendMetrics(p, c, 0)

	return
} // Get()

// `Metrics()` returns the current pool metrics.
//
// The returned metrics are a snapshot of the current state at the time
// of calling this method. The metrics may change as soon as the method
// returns.
//
// The returned metrics show:
//   - `Size`: Current number of nodes in the pool.
//   - `Created`: Number of nodes created by the pool.
//   - `Returned`: Number of nodes returned to the pool.
//
// Returns:
//   - `rMetric`: Current pool metrics.
//   - `rErr`: An error, if any.
func (p *TPool) Metrics() (rMetric *TPoolMetrics, rErr error) {
	if nil == p {
		rErr = ErrPoolNotInit
		return
	}
	if nil == p.nodes {
		p.reset(0)
	}
	rMetric = &TPoolMetrics{
		Size:     len(p.nodes),
		Created:  p.created.Load(),
		Returned: p.returned.Load(),
	}

	return
} // Metrics()

// `MetricsChannel()` returns a R/O channel that provides the current pool
// metrics.
//
// The channel can be used for real-time monitoring of the pool's activity.
//
// The metrics returned by the channel show:
//   - `Size`: Current number of nodes in the pool.
//   - `Created`: Number of nodes created by the pool.
//   - `Returned`: Number of nodes returned to the pool.
//
// Returns:
//   - `rChan`: Channel that provides the pool metrics.
//   - `rErr`: An error, if any.
func (p *TPool) MetricsChannel() (rChan <-chan *TPoolMetrics, rErr error) {
	if nil == p {
		rErr = ErrPoolNotInit
		return
	}
	if nil == p.mCh {
		// Pool not initialised yet
		p.reset(0)
	}
	rChan = p.mCh

	return
} // MetricsChannel()

// `newNode()` creates a new node.
//
// This private method is called internally by the `Get()` method if the pool
// is empty. It utilises the pool's `New()` function to create the new node.
//
// Returns:
//   - `rNode`: A new node.
func (p *TPool) newNode() (rNode any, rCreated uint32) {
	if nil == p.New {
		// Factory function not set: just return the default value
		return
	}

	rNode = p.New()
	rCreated = p.created.Add(1)
	//TODO: Go 1.24:
	// runtime.AddCleanup(rNode, func() {
	// 	p.put(rNode)
	// })

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
	if nil == p.nodes {
		// Pool not initialised yet
		p.reset(0)
	}

	r := p.returned.Add(1)
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
	p.mCh = make(chan *TPoolMetrics, 1) // pool metrics

	// Pre-allocate some nodes for the pool:
	if nil != p.New {
		for range aSize {
			p.Put(p.New()) //#nosec G104 -- ignore the (here impossible) error
		}
	}
} // reset()

// `sendMetrics()` sends the pool's metrics to the metrics channel.
//
// This function is called asynchronously by the `Get()` and `Put()` methods.
//
// Parameters:
//   - `aPool`: The pool to send the metrics for.
//   - `aCreate`: Number of nodes created (if `0`, use pool's counter).
//   - `aReturn`: Number of nodes returned (if `0`, use pool's counter).
func sendMetrics(aPool *TPool, aCreate, aReturn uint32) {
	if 0 == aCreate {
		aCreate = aPool.created.Load()
	}
	if 0 == aReturn {
		aReturn = aPool.returned.Load()
	}
	m := &TPoolMetrics{
		Size:     len(aPool.nodes),
		Created:  aCreate,
		Returned: aReturn,
	}

	select {
	case aPool.mCh <- m:
		// Metrics were written
		runtime.Gosched()
	case <-time.After(time.Second):
		// Timeout (just to be sure to not block)
		return
	default:
		// Ignore if nobody's listening
	}
} // sendMetrics()

/* _EoF_ */
