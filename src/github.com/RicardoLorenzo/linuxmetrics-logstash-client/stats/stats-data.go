package stats

import (
  "fmt"
  "log"
  "os"
  "sync"
  "time"
  "strconv"
  "bytes"
  "io/ioutil"
  "strings"

  linuxproc "github.com/c9s/goprocinfo/linux"
)

var (
  ProcPath string
  SharedStatsPeriod StatsPeriod = NewStatsPeriod()
  SignalChannel chan uint32 = make(chan uint32)
)

type StatsSample struct {
  time uint64
  hostname string
  stat *linuxproc.Stat
  cpuinfo *linuxproc.CPUInfo
  vmstat *linuxproc.VMStat
  snmp *linuxproc.Snmp
  tcpsockets []*linuxproc.NetTCPSocket
  meminfo *linuxproc.MemInfo
  processes []*linuxproc.Process
  diskstats []*linuxproc.DiskStat
}

type StatsPeriod struct {
  previous StatsSample
  current StatsSample
  rwlock sync.RWMutex
}

func (statsSample *StatsSample) getHostname(procPath string) string {
  var path string = procPath + "sys/kernel/hostname"

  data, err := ioutil.ReadFile(path)
  if err != nil {
		return "unknown"
	}
  return strings.TrimSuffix(string(data), "\n")
}

func (statsSample *StatsSample) getProcessPath(pid uint64, file string) string {
  var buffer bytes.Buffer

  buffer.WriteString(ProcPath)
  buffer.WriteString(strconv.FormatUint(pid, 10))
  if file != "" {
    buffer.WriteString("/")
    buffer.WriteString(file)
  }
  return buffer.String()
}

func (statsSample *StatsSample) getProcessesIds(path string) ([]uint64, error) {
	var pids []uint64

	procDirectory, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer procDirectory.Close()

	allSubDirectories, err := procDirectory.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	for _, fname := range allSubDirectories {
    var pid uint64
    pid, err := strconv.ParseUint(fname, 10, 32)
		if err != nil {
			// Not numeric value
			continue
		}
    /**
     * This is based on the assumption that user level processes are
     * always backed by a binary file in /proc/<pid>/exe
     */
     var path string = statsSample.getProcessPath(pid, "exe")
     if _, err := os.Stat(path); err == nil {
       if !os.IsNotExist(err) {
         pids = append(pids, uint64(pid))
       }
     }
	}

	return pids, nil
}

func NewStatsSample() (statsSample StatsSample, err error) {
  statsSample.hostname = statsSample.getHostname(ProcPath)
  statsSample.stat, err = linuxproc.ReadStat(ProcPath + "stat")
  if err != nil {
  	return statsSample, err
  }
  statsSample.cpuinfo, err = linuxproc.ReadCPUInfo(ProcPath + "cpuinfo")
  if err != nil {
  	return statsSample, err
  }
  statsSample.vmstat, err = linuxproc.ReadVMStat(ProcPath + "vmstat")
  if err != nil {
  	return statsSample, err
  }
  statsSample.snmp, err = linuxproc.ReadSnmp(ProcPath + "net/snmp")
  if err != nil {
  	return statsSample, err
  }

  sockets, err := linuxproc.ReadNetTCPSockets(ProcPath + "net/tcp", linuxproc.NetIPv4Decoder)
  if err != nil {
  	return statsSample, err
  }

  // Without accessing by index, object references are incorrect
  for i, _ := range sockets.Sockets {
    statsSample.tcpsockets = append(statsSample.tcpsockets, &sockets.Sockets[i])
  }

  disks, err := linuxproc.ReadDiskStats(ProcPath + "diskstats")
  if err != nil {
  	return statsSample, err
  }
  // Without accessing by index, object references are incorrect
  for i, _ := range disks {
    statsSample.diskstats = append(statsSample.diskstats, &disks[i])
  }

  /**
   * Getting list of all user-level processes
   */
  processesIds, err := statsSample.getProcessesIds(ProcPath)
  if err != nil {
  	return statsSample, err
  }

  for _, pid := range processesIds {
    process, err := linuxproc.ReadProcess(pid, ProcPath)
    if err != nil {
    	return statsSample, err
    }
    statsSample.processes = append(statsSample.processes, process)
  }

  statsSample.time = uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
	return statsSample, nil
}

func (statsSample *StatsSample) getSystemCPUHertz() float64 {
  return statsSample.cpuinfo.Processors[0].MHz * 1024 * 1024
}

func (statsSample *StatsSample) getSystemCPUCount() uint64 {
  return uint64(len(statsSample.cpuinfo.Processors))
}

func (statsSample *StatsSample) clone() StatsSample {
    clone := *statsSample
    return clone
}

func NewStatsPeriod() StatsPeriod {
  statsPeriod := StatsPeriod{}
	return statsPeriod
}

func (statsPeriod *StatsPeriod) AddStatsSample(stats StatsSample) {
	statsPeriod.rwlock.Lock()
  defer statsPeriod.rwlock.Unlock()
  statsPeriod.previous = statsPeriod.current.clone()
  statsPeriod.current = stats.clone()
}

func (statsPeriod *StatsPeriod) GetStatsSamples() (previous, current StatsSample) {
  statsPeriod.WaitForSamples()
	statsPeriod.rwlock.RLock()
	defer statsPeriod.rwlock.RUnlock()
  return statsPeriod.previous, statsPeriod.current
}

func (statsPeriod *StatsPeriod) HasPreviousSamples() bool {
  if statsPeriod.previous.time != 0 {
    return true
  }
  return false
}

func (statsPeriod *StatsPeriod) WaitForSamples() {
  if ! statsPeriod.HasPreviousSamples() {
    for {
      receivedCode := <-SignalChannel
      if receivedCode == 1 {
        break
      }
    }
  }
}

func CollectStatsSamples(secondsInterval time.Duration) {
  for {
    statsSample, err := NewStatsSample()
    if err != nil {
      log.Println(fmt.Sprint(err), err)
    }
    SharedStatsPeriod.AddStatsSample(statsSample)
    if SharedStatsPeriod.HasPreviousSamples() {
      /**
       * Non-blocking approach.
       * This is not a fancy thing, it is important.
       */
      select {
        case SignalChannel <- 1:
            // signal sent
        default:
            // signal dropped
      }
    }
    time.Sleep(secondsInterval * time.Second)
  }
}
