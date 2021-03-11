# nrradix
[![Go Reference](https://pkg.go.dev/badge/github.com/nobuyo/nrradix.svg)](https://pkg.go.dev/github.com/nobuyo/nrradix)

A wrapper with New Relic instrumentation for [mediocregopher/radix](https://github.com/mediocregopher/radix) v3.


## Getting Started

```go
import "github.com/nobuyo/nrradix"
```


### Run Simple Command

```go
txn := ... // Your *newrelic.Transaction
pool, err := nrradix.NewPool("tcp", url, poolSize)
if nil != err {
    fmt.Println(err)
    os.Exit(1)
}

pool.Do(txn, nil, "SET", 1)
```

### Run Pipeline

```go
var commands []nrradix.CmdElement

commands = append(commands, nrradix.NewCmdElement(nil, "SET", "FOO"))
commands = append(commands, nrradix.NewCmdElement(nil, "SET", "BAR"))

pool.DoPipeline(txn, commands)
```

