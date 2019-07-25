package main

import (
  "fmt"
  "flag"
  "log"
	"os"
  "strconv"
  "time"
	"os/signal"
	"syscall"
  "strings"

  config "github.com/RicardoLorenzo/linuxmetrics-logstash-client/config"
  stats "github.com/RicardoLorenzo/linuxmetrics-logstash-client/stats"
  logstash "github.com/RicardoLorenzo/linuxmetrics-logstash-client/logstash"
)

const (
  // Interval for sampling collection and process
  defaultMillisInterval int = 1
  defaultHost string = "127.0.0.1"
  defaultPort int = 1514
)

var configPath string
var logstashHost string
var logstashPort int
var secondsInterval int

func init() {
    if flag.Lookup("c") == nil {
        flag.StringVar(&configPath, "c", "", "Configuration file")
    }
    if flag.Lookup("host") == nil {
        flag.StringVar(&logstashHost, "host", "", "Logstash hostname")
    }
    if flag.Lookup("port") == nil {
        flag.IntVar(&logstashPort, "port", -1, "Logstash port")
    }
    if flag.Lookup("interval") == nil {
        flag.IntVar(&secondsInterval, "interval", defaultMillisInterval, "Seconds between samples")
    }
    if flag.Lookup("proc-path") == nil {
        flag.StringVar(&stats.ProcPath, "proc-path", "", "Linux proc path")
    }
    if flag.Lookup("console") == nil {
        flag.BoolVar(&logstash.ConsoleOutput, "console", false, "Console output")
    }
}

func main() {
  flag.Parse()
  configPath = flag.Lookup("c").Value.(flag.Getter).Get().(string)
  logstashHost = flag.Lookup("host").Value.(flag.Getter).Get().(string)
  logstashPort = flag.Lookup("port").Value.(flag.Getter).Get().(int)
  secondsInterval = flag.Lookup("interval").Value.(flag.Getter).Get().(int)
  stats.ProcPath = flag.Lookup("proc-path").Value.(flag.Getter).Get().(string)
  logstash.ConsoleOutput = flag.Lookup("console").Value.(flag.Getter).Get().(bool)

  /**
   * Creates a channel and waits for SIGTERM to exit application
   */
  channel := make(chan os.Signal, 2)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-channel
		os.Exit(1)
	}()

  log.SetOutput(os.Stdout)
  config, err := config.NewConfig(configPath)
  	if err != nil {
  		log.Panic(fmt.Sprint(err), err)
  }

  if logstashHost == "" {
    logstashHost = config.GetProperty("logstash.hostname", defaultHost)
  }
  if logstashPort == -1 {
    logstashPort, err = strconv.Atoi(config.GetProperty("logstash.port",
      strconv.Itoa(defaultPort)))
    if err != nil {
      logstashPort = defaultPort
    }
  }
  if stats.ProcPath == "" {
    stats.ProcPath = config.GetProperty("proc.path", "/proc")
  }

  if(!strings.HasSuffix(stats.ProcPath, "/")) {
    stats.ProcPath = stats.ProcPath + "/"
  }

  logstash := logstash.NewLogstashClient(logstashHost, logstashPort, 5000)

  // This background thread collects the samples from the OS
  go stats.CollectStatsSamples(time.Duration(secondsInterval))
  /**
  * This background thread reads the samples from the event channel
  * and send them to Logstash
  */
  go logstash.ReadEventsFromBacklog()

  jsonstats := stats.NewJSONStats()
  for {
    eventMessage, err := jsonstats.GetStats()
    if err != nil {
  		log.Panic("Statistics collection error: ", fmt.Sprint(err), err)
    }
    logstash.SendEventToBacklog(eventMessage)
    time.Sleep(time.Duration(secondsInterval) * time.Second)
  }
}
