# skycoin-services
skycon-services

## Dmsg-daemon
To generate and register dmsg-daemon use:
```shell
$ make daemon
```

To start dmsg-daemon use:
```shell
$ make daemon-start
```

To check the daemon logs use: 
```shell
$ journalctl -f -u dmsg-daemon
```

To stop dmsg-daemon use:
```shell
$ make daemon-stop
```

To start dmsg-daemon on reoot use:
```shell
$ make deamon-start-on-reboot
```