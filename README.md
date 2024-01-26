# nomad-check

Check your nomad

```
go get github.com/schmichael/nomad-check
nomad-check -h # for help
nomad-check    # to run against the http api

# to run against nomad debug dumps
nomad-check -allocs allocations.json -nodes nodes.json
```

Checkout nomad-check.json for results
