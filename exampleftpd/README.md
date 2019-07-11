# exampleftpd

This is a simple example ftpd server for testing against and to demonstrate how to use the interface.

Change to this directory then build it with `go build`

```
$ ./exampleftpd -h
Usage of ./exampleftpd:
  -host string
    	Host (default "localhost")
  -pass string
    	Password for login (default "123456")
  -port int
    	Port (default 2121)
  -root string
    	Root directory to serve
  -user string
    	Username for login (default "admin")
```
