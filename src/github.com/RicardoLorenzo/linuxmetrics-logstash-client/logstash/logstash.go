package logstash

import (
	"net"
	"fmt"
	"log"
	"io"
	"time"
	"errors"
)

var (
  EventChannel chan string = make(chan string, 20)
	ConsoleOutput bool
)

type LogstashClient struct {
	Hostname string
	Port int
	Connection *net.TCPConn
	SocketTimeout int
}

func NewLogstashClient(hostname string, port int, socketTimeoutMS int) *LogstashClient {
	logstash := LogstashClient{}
	logstash.Hostname = hostname
	logstash.Port = port
	logstash.Connection = nil
	logstash.SocketTimeout = socketTimeoutMS
	return &logstash
}

func (logstash *LogstashClient) setConnectionDeadline() {
	timeInMillis := time.Now().Add(time.Duration(logstash.SocketTimeout) * time.Millisecond)
	logstash.Connection.SetDeadline(timeInMillis)
	logstash.Connection.SetWriteDeadline(timeInMillis)
	logstash.Connection.SetReadDeadline(timeInMillis)
}

func (logstash *LogstashClient) Connect() (*net.TCPConn, error) {
  var connection *net.TCPConn
	service := fmt.Sprintf("%s:%d", logstash.Hostname, logstash.Port)
	addr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		return connection, err
	}
	if(addr == nil) {
		log.Panic("Logstash host [", logstash.Hostname, "] connot be resolved")
	}
	connection, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		return connection, err
	}
	if connection != nil {
		logstash.Connection = connection
		logstash.Connection.SetLinger(0) // default -1
		logstash.Connection.SetNoDelay(true)
		logstash.Connection.SetKeepAlive(true)
		logstash.Connection.SetKeepAlivePeriod(time.Duration(5) * time.Second)
		logstash.setConnectionDeadline()
	}
	return connection, err
}

func (logstash *LogstashClient) reConnect() {
	for {
		time.Sleep(500 * time.Millisecond)
		log.Println("=>Reconnecting ...")
		if logstash.Connection != nil {
			logstash.Connection.Close()
			logstash.Connection = nil
		}
		_, err := logstash.Connect()
		if _, ok := err.(net.Error); ok {
			log.Println("Logstash client connection attmept has failed - ", fmt.Sprint(err))
		} else if err != nil {
			log.Panic("Logstash client connection cannot be re-established", fmt.Sprint(err), err)
		} else {
			log.Println("=>Connection re-established")
			break
		}
	}
}

func (logstash *LogstashClient) Send(message string) (error) {
	var err = errors.New("TCP Connection is nil.")
	message = fmt.Sprintf("%s\n", message)
	if logstash.Connection != nil {
		for {
			_, err = logstash.Connection.Write([]byte(message))
			if err != nil {
				if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					log.Println("Logstash client socket timeout from ", logstash.Connection.RemoteAddr())
					log.Println("=>Message discarded")
					// Autohealing attempt
					logstash.reConnect()
					return err
		    } else if err == io.EOF {
					log.Println("Logstash client disconnected from ", logstash.Connection.RemoteAddr())
					log.Println("=>Retrying")
					// Autohealing attempt
          logstash.reConnect()
					continue
				} else {
					log.Println("Logstash client connection error - ", fmt.Sprint(err), err)
					log.Println("=>Message discarded")
					// Autohealing attempt
					logstash.reConnect()
					return err
				}
			} else {
	      // Sets the deadline for future Write/Read calls.
				logstash.setConnectionDeadline()
				return nil
			}
		}
	}
	return err
}

func (logstash *LogstashClient) ReadEventsFromBacklog() {
	_, err := logstash.Connect()
	if _, ok := err.(net.Error); ok {
		log.Println("Logstash client connection attmept has failed - ", fmt.Sprint(err))
		logstash.reConnect()
	} else if err != nil {
		log.Panic("Logstash client connection cannot be established - ", fmt.Sprint(err), err)
	}

	for {
		receivedEvent := <-EventChannel

		if(ConsoleOutput) {
			fmt.Println(" ---")
			fmt.Println(receivedEvent)
		}
	  logstash.Send(receivedEvent)
	}
}

func (logstash *LogstashClient) SendEventToBacklog(message string) {
	EventChannel <- message
}
