package nrradix

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/mediocregopher/radix/v3"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
)

// WrappedPool ...
type WrappedPool struct {
	*radix.Pool
	redisURL *url.URL
}

// NewPool wraps radix.NewPool
func NewPool(network, addr string, size int, opts ...radix.PoolOpt) (*WrappedPool, error) {
	pool, err := radix.NewPool(network, addr, size, opts...)
	if nil != err {
		return nil, err
	}

	url, err := url.Parse(addr)
	if nil != err {
		return nil, err
	}

	return &WrappedPool{
		Pool:     pool,
		redisURL: url,
	}, nil
}

// Do wraps radix.Pool.Do()
func (p *WrappedPool) Do(txn *newrelic.Transaction, rcv interface{}, cmd string, args ...string) error {
	cmdAction := radix.Cmd(rcv, cmd, args...)
	if txn != nil {
		seg := p.newSegment(txn, cmd, cmdAction)
		defer seg.End()
	}

	return p.Pool.Do(cmdAction)
}

func (p *WrappedPool) newSegment(txn *newrelic.Transaction, commandName string, cmdAction radix.CmdAction) *newrelic.DatastoreSegment {
	segment := newrelic.DatastoreSegment{}
	segment.Product = newrelic.DatastoreRedis
	segment.Operation = strings.ToLower(commandName)
	segment.StartTime = txn.StartSegmentNow()
	segment.Host = p.redisURL.Host
	segment.PortPathOrID = p.redisURL.Port()
	segment.ParameterizedQuery = fmt.Sprint(cmdAction)

	return &segment
}
