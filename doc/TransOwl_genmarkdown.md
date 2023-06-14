## TransOwl genmarkdown

Generate Instruction

### Synopsis

create markdown at location specified

```
TransOwl genmarkdown [flags]
```

### Examples

```
./TransOwl genmarkdown --mdpath ./doc
```

### Options

```
  -h, --help            help for genmarkdown
      --mdpath string   Create markdown at specified location
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

