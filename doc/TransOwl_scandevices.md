## TransOwl scandevices

Print all devices available in current net.

### Synopsis

Only devices responding TransOwl UDP packet are accepted.

```
TransOwl scandevices [flags]
```

### Examples

```
./TransOwl scandevices
```

### Options

```
  -h, --help   help for scandevices
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

