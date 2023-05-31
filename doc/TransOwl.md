## TransOwl

TransOwl

### Synopsis

TransOwl - A simple tool for file transition

```
TransOwl [flags]
```

### Options

```
  -d, --deepscan                Scan in 255.255.255.255; If not specified, devices with the same network segment as the NIC are scanned.
  -h, --help                    help for TransOwl
  -i, --interface stringArray   Specify interface you want to search devices in
      --logtofile string        Specify a location logs storage in, default is ./TransOwl_*.log
  -u, --user string             Specify a username
      --verbose                 Print Debug Level logs
      --vverbose                Print Debug/Trace Level logs
```

### SEE ALSO

* [TransOwl completion](TransOwl_completion.md)	 - Generate the autocompletion script for the specified shell
* [TransOwl genmarkdown](TransOwl_genmarkdown.md)	 - Generate Instruction
* [TransOwl netls](TransOwl_netls.md)	 - List net available
* [TransOwl scandevices](TransOwl_scandevices.md)	 - Print all devices available in current net.
* [TransOwl sendfile](TransOwl_sendfile.md)	 - send file to someone
* [TransOwl waitrecv](TransOwl_waitrecv.md)	 - Wait for receiving file

