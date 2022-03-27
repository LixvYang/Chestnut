# Heterogeneous Solutions for the Internet.

![](chestnut_arch.png)

## Build
```go
go build
```

## Help
```
Output a help 

Useage...
  -alsologtostderr
        log to standard error as well as files
  -apilisten string
        Adds a multiaddress to the listen list (default ":5215")
  -bootstrap
        run a bootstrap node
  -configdir string
        config and keys dir (default "./config/")
  -datadir string
        config dir (default "./data/")
  -debug
        show debug log
  -h    Display help
  -ips value
        IPAddresses field of x509 certificate
  -jsontracer string
        output tracer data to a json file
  -keystoredir string
        keystore dir (default "./keystore/")
  -keystorename string
        keystore name (default "defaultkeystore")
  -listen -listen /ip4/127.0.0.1/tcp/4215 -listen /ip/127.0.0.1/tcp/5215/ws
        Adds a multiaddress to the listen list, e.g.: -listen /ip4/127.0.0.1/tcp/4215 -listen /ip/127.0.0.1/tcp/5215/ws        
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -log_dir string
        If non-empty, write log files in this directory
  -logtostderr
        log to standard error instead of files
  -peer value
        Adds a peer multiaddress to the bootstrap list
  -peername string
        peername (default "peer")
  -ping
        ping peer
  -rendezvous string
        Unique string to identify group of nodes. Share this with your friends to let them connect with you (default "e6629921-b5cd-4855-9fcd-08bcc39caef7")
  -stderrthreshold value
        logs at or above this threshold go to stderr
  -v value
        log level for V logs
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging
```

The project is in the process of making.


Expecting...