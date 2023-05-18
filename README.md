# Lsf Exporter 

[Prometheus](https://prometheus.io/) exporter for IBM Spectrum LSF Manager


## Install

```shell
$ go install github.com/a270443177/lsf_exporter
```

## Building

```shell
$ cd $GOPATH/src/github.com/a270443177/lsf_exporter
$ make
```

## Configuration


Notes:

## Running

 1. source LSF profile file
    ```
    bash:
    $ source <LSF_TOP>/conf/profile.lsf
    OR
    csh:
    $ source <LSF_TOP>/conf/cshrc.lsf 

    ```

 2. run lsf_exporter 
    ```

    $ ./flexlm_exporter <flags>

    ```


Metrics will now be reachable at http://localhost:9818/metrics.

## What's exported?

 * `lsid` information.
 * `bhosts -w` bhosts information.
 * `bqueues -w`  bqueues information.

