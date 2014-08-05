package server

import (
	"bufio"
	"crypto/tls"
	"net"
	"strconv"
	"strings"
)

// serverOpts contains parameters for graval.NewServer()
type ServerOpts struct {
	// The factory that will be used to create a new FTPDriver instance for
	// each client connection. This is a mandatory option.
	Factory DriverFactory

	Auth Auth

	// Server Name, Default is Go Ftp Server
	Name string

	// The hostname that the FTP server should listen on. Optional, defaults to
	// "::", which means all hostnames on ipv4 and ipv6.
	Hostname string

	// The port that the FTP should listen on. Optional, defaults to 3000. In
	// a production environment you will probably want to change this to 21.
	Port int

	// use tls, default is false
	TLS bool

	// if tls used, cert file is required
	CertFile string

	// if tls used, key file is required
	KeyFile string

	WelcomeMessage string
}

// Server is the root of your FTP application. You should instantiate one
// of these and call ListenAndServe() to start accepting client connections.
//
// Always use the NewServer() method to create a new Server.
type Server struct {
	*ServerOpts
	name          string
	listenTo      string
	driverFactory DriverFactory
	logger        *Logger
}

// serverOptsWithDefaults copies an ServerOpts struct into a new struct,
// then adds any default values that are missing and returns the new data.
func serverOptsWithDefaults(opts *ServerOpts) *ServerOpts {
	var newOpts ServerOpts
	if opts == nil {
		opts = &ServerOpts{}
	}
	if opts.Hostname == "" {
		newOpts.Hostname = "::"
	} else {
		newOpts.Hostname = opts.Hostname
	}
	if opts.Port == 0 {
		newOpts.Port = 3000
	} else {
		newOpts.Port = opts.Port
	}
	newOpts.Factory = opts.Factory
	if opts.Name == "" {
		newOpts.Name = "Go FTP Server"
	} else {
		newOpts.Name = opts.Name
	}

	if opts.WelcomeMessage == "" {
		newOpts.WelcomeMessage = defaultWelcomeMessage
	} else {
		newOpts.WelcomeMessage = opts.WelcomeMessage
	}

	if opts.Auth == nil {
		newOpts.Auth = AnonymousAuth{}
	} else {
		newOpts.Auth = opts.Auth
	}

	newOpts.TLS = opts.TLS
	newOpts.KeyFile = opts.KeyFile
	newOpts.CertFile = opts.CertFile

	return &newOpts
}

// NewServer initialises a new FTP server. Configuration options are provided
// via an instance of ServerOpts. Calling this function in your code will
// probably look something like this:
//
//     factory := &MyDriverFactory{}
//     server  := graval.NewServer(&graval.ServerOpts{ Factory: factory })
//
// or:
//
//     factory := &MyDriverFactory{}
//     opts    := &graval.ServerOpts{
//       Factory: factory,
//       Port: 2000,
//       Hostname: "127.0.0.1",
//     }
//     server  := graval.NewServer(opts)
//
func NewServer(opts *ServerOpts) *Server {
	opts = serverOptsWithDefaults(opts)
	s := new(Server)
	s.ServerOpts = opts
	s.listenTo = buildTcpString(opts.Hostname, opts.Port)
	s.name = opts.Name
	s.driverFactory = opts.Factory
	s.logger = newLogger("")
	return s
}

// NewConn constructs a new object that will handle the FTP protocol over
// an active net.TCPConn. The TCP connection should already be open before
// it is handed to this functions. driver is an instance of FTPDriver that
// will handle all auth and persistence details.
func (server *Server) newConn(tcpConn net.Conn, driver Driver) *Conn {
	c := new(Conn)
	c.namePrefix = "/"
	c.conn = tcpConn
	c.controlReader = bufio.NewReader(tcpConn)
	c.controlWriter = bufio.NewWriter(tcpConn)
	c.driver = driver
	c.server = server
	c.sessionId = newSessionId()
	c.logger = newLogger(c.sessionId)
	return c
}

func simpleTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	config := &tls.Config{}
	if config.NextProtos == nil {
		config.NextProtos = []string{"ftp"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// ListenAndServe asks a new Server to begin accepting client connections. It
// accepts no arguments - all configuration is provided via the NewServer
// function.
//
// If the server fails to start for any reason, an error will be returned. Common
// errors are trying to bind to a privileged port or something else is already
// listening on the same port.
//
func (Server *Server) ListenAndServe() error {
	/*laddr, err := net.ResolveTCPAddr("tcp", Server.listenTo)
	if err != nil {
		return err
	}*/

	var listener net.Listener
	var err error
	//fmt.Println("-------", *Server.ServerOpts)
	if Server.ServerOpts.TLS {
		//fmt.Println("use tls")
		config, err := simpleTLSConfig(Server.CertFile, Server.KeyFile)
		if err != nil {
			return err
		}

		listener, err = tls.Listen("tcp", Server.listenTo, config)
	} else {
		listener, err = net.Listen("tcp", Server.listenTo)
	}
	if err != nil {
		return err
	}

	Server.logger.Printf("%s listening on %d", Server.Name, Server.Port)

	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			Server.logger.Print("listening error")
			break
		}
		driver, err := Server.driverFactory.NewDriver()
		if err != nil {
			Server.logger.Print("Error creating driver, aborting client connection")
		} else {
			ftpConn := Server.newConn(tcpConn, driver)
			go ftpConn.Serve()
		}
	}
	return nil
}

func buildTcpString(hostname string, port int) (result string) {
	if strings.Contains(hostname, ":") {
		// ipv6
		if port == 0 {
			result = "[" + hostname + "]"
		} else {
			result = "[" + hostname + "]:" + strconv.Itoa(port)
		}
	} else {
		// ipv4
		if port == 0 {
			result = hostname
		} else {
			result = hostname + ":" + strconv.Itoa(port)
		}
	}
	return
}
