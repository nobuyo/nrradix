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

// CmdElement ...
type CmdElement struct {
	rcv  interface{}
	cmd  string
	args []string
}

// NewCmdElement creates CmdElement as like original Do() interface
func NewCmdElement(rcv interface{}, cmd string, args ...string) CmdElement {
	return CmdElement{
		rcv:  rcv,
		cmd:  cmd,
		args: args,
	}
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

// DoPipeline wraps radix.Pool.Do() with Pipeline
func (p *WrappedPool) DoPipeline(txn *newrelic.Transaction, cmds []CmdElement) error {
	pipeline := radix.Pipeline(generateCommands(cmds)...)
	if txn != nil {
		seg := p.newPipelineSegment(txn, operationName(cmds), operationString(cmds))
		defer seg.End()
	}

	return p.Pool.Do(pipeline)
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

func (p *WrappedPool) newPipelineSegment(txn *newrelic.Transaction, commandName string, operation string) *newrelic.DatastoreSegment {
	segment := newrelic.DatastoreSegment{}
	segment.Product = newrelic.DatastoreRedis
	segment.Operation = strings.ToLower(commandName)
	segment.StartTime = txn.StartSegmentNow()
	segment.Host = p.redisURL.Host
	segment.PortPathOrID = p.redisURL.Port()
	segment.ParameterizedQuery = operation

	return &segment
}

func operationName(cmdElements []CmdElement) string {
	operations := make([]string, len(cmdElements))
	for n := range cmdElements {
		operations[n] = cmdElements[n].cmd
	}

	return "pipeline: " + strings.Join(operations, ", ")
}

func operationString(cmdElements []CmdElement) string {
	operations := make([]string, len(cmdElements))
	for n := range cmdElements {
		operations[n] = cmdElements[n].cmd + " " + strings.Join(cmdElements[n].args, " ")
	}

	return strings.Join(operations, "; ")
}

func generateCommands(cmdElements []CmdElement) []radix.CmdAction {
	commands := make([]radix.CmdAction, len(cmdElements))
	for n := range cmdElements {
		element := cmdElements[n]
		commands[n] = radix.Cmd(element.rcv, element.cmd, element.args...)
	}

	return commands
}
