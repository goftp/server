# raval

An experimental FTP server framework. By providing a simple driver class that
responds to a handful of methods you can have a complete FTP server.

Some sample use cases include persisting data to:

* an Amazon S3 bucket
* a relational database
* redis
* memory

There is a sample in-memory driver available on github:

* [https://github.com/yob/graval-mem](https://github.com/yob/graval-mem)

## Installation

    go get github.com/yob/graval

## Usage

To boot an FTP server you will need to provide a driver that speaks to your
persistence layer - the required driver contract is listed below.

Once that's ready, boot a new server like so:

    package main

    import "github.com/yob/graval"

    type MemDriver struct {}

    .. implement MemDriver here ..

    type MemDriverFactory struct {}

    func (factory *MemDriverFactory) NewDriver() graval.FTPDriver {
      return &MemDriver{}
    }

    func main() {
      factory := &MemDriverFactory{}
      ftpServer := graval.NewFTPServer("localhost", 3000, factory)
      ftpServer.Listen()
    }

### The Driver Contract

The driver MUST have the following methods.  Each method MUST accept the listed
parameters and return an appropriate value:

    authenticate(user, pass)
    - boolean indicating if the provided details are valid

## Contributors

* James Healy <james@yob.id.au> [http://www.yob.id.au](http://www.yob.id.au)

## Warning

FTP is an incredibly insecure protocol. Be careful about forcing users to authenticate
with a username or password that are important.

## License

This library is distributed under the terms of the MIT License. See the included file for
more detail.

## Contributing

All suggestions and patches welcome, preferably via a git repository I can pull from.
If this library proves useful to you, please let me know.

## Further Reading

There are a range of RFCs that together specify the FTP protocol. In chronological
order, the more useful ones are:

* [http://tools.ietf.org/rfc/rfc959.txt](http://tools.ietf.org/rfc/rfc959.txt)
* [http://tools.ietf.org/rfc/rfc1123.txt](http://tools.ietf.org/rfc/rfc1123.txt)
* [http://tools.ietf.org/rfc/rfc2228.txt](http://tools.ietf.org/rfc/rfc2228.txt)
* [http://tools.ietf.org/rfc/rfc2389.txt](http://tools.ietf.org/rfc/rfc2389.txt)
* [http://tools.ietf.org/rfc/rfc2428.txt](http://tools.ietf.org/rfc/rfc2428.txt)
* [http://tools.ietf.org/rfc/rfc3659.txt](http://tools.ietf.org/rfc/rfc3659.txt)
* [http://tools.ietf.org/rfc/rfc4217.txt](http://tools.ietf.org/rfc/rfc4217.txt)

For an english summary that's somewhat more legible than the RFCs, and provides
some commentary on what features are actually useful or relevant 24 years after
RFC959 was published:

* [http://cr.yp.to/ftp.html](http://cr.yp.to/ftp.html)

For a history lesson, check out Appendix III of RCF959. It lists the preceding
(obsolete) RFC documents that relate to file transfers, including the ye old
RFC114 from 1971, "A File Transfer Protocol"

This library is heavily based on [em-ftpd](https://github.com/yob/em-ftpd), an FTPd
framework with similar design goals within the ruby and EventMachine ecosystems. It
worked well enough, but you know, callbacks and event loops make me something
something.
