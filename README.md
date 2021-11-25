tcpterm
===

tcpterm visualize packets in TUI.

Demo
---

[![asciicast](https://asciinema.org/a/td3DA8LH04XYhxGPirJvsEI4V.png)](https://asciinema.org/a/td3DA8LH04XYhxGPirJvsEI4V)

Install
---

*Requires libpacp header files for compilation*

For Debian based distriubutions install package _libpcap-dev_

```
$ go install github.com/sachaos/tcpterm
```

Usage
---

```
$ tcpterm -h
NAME:
   tcpterm - tcpdump for human

USAGE:
   tcpterm [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --interface value, -i value  If unspecified, use lowest numbered interface.
   --read value, -r value       Read packets from pcap file.
   --filter value, -f value     BPF Filter
   --debug                      debug mode.
   --help, -h                   show help
   --version, -v                print the version
```

TODO
---

* Optimize packets list view.
* Improve detail view, and dump view.
