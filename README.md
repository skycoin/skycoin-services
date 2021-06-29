# skycoin-services

skycon-services

## dmsg daemon

This daemon will create `.dmsg-uuid` file, which contains your unique UUID.  
It also log all the actions and results to the `dmsg-uuid.log` file.  
These files will be created in the same directory, where **dmsgd** are running.

- **port** - to start HTTP server on
- **disc** - dmsg discovery server
- **sk** - dmsg server secret key
- **log-dir** - directory to store logs and UUID data

### Building from source

1. `go build -o dmsgd ./dmsg`
2. `./dmsgd`

### Running in Docker

1. `docker build -t dmsgd . && docker run -d -v /$PWD/docker_container:/dmsgd-data dmsgd --port=80 --disc="http://dmsg.discovery.skywire.skycoin.com" --sk=***`