## TransOwl waitrecv

Wait for receiving file from host

### Synopsis

This exits after300seconds

```
TransOwl waitrecv [flags]
```

### Examples

```
./TransOwl waitrecv -u TransOwl --savepath /tmp/transowl
```

### Options

```
  -h, --help              help for waitrecv
      --savepath string   file will be saved at path you specified.
```

### Options inherited from parent commands

```
  -d, --deepscan                Scan in 255.255.255.255; If not specified, devices with the same network segment as the NIC are scanned.
  -i, --interface stringArray   Specify interface you want to search devices in
      --logtofile string        Specify a location logs storage in, default is ./TransOwl_*.log
  -u, --user string             Specify a username
      --verbose                 Print Debug Level logs
      --vverbose                Print Debug/Trace Level logs
```

### SEE ALSO

* [TransOwl](TransOwl.md)	 - TransOwl

