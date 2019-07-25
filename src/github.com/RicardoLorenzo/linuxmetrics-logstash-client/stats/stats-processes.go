package stats

import (
	"math"
)

type LinuxProcessStats struct {
	CmdLine string `json:"cmdline"`
	Pid uint64 `json:"pid"`
	State string `json:"state"`
	MemVirtSize uint64 `json:"mem_virtual_size"`
	MemRssSize uint64 `json:"mem_rss_size"`
	MemLockSize uint64 `json:"mem_lock_size"`
	MemSwapSize uint64 `json:"mem_swap_size"`
	Threads uint64 `json:"threads"`
	FDUsed uint64 `json:"fd_used"`
	SignalsIgnored uint64 `json:"sig_ignored"`
	SignalsCaught uint64 `json:"sig_caught"`
	VoluntaryContextSwitches uint64 `json:"voluntary_contextswitches"`
	NonVoluntaryContextSwitches uint64 `json:"nonvoluntary_contextswitches"`
	IOReadBytes uint64 `json:"io_read_bytes"`
	IOWriteBytes uint64 `json:"io_write_bytes"`
	UserCpuUsage uint64 `json:"user_cpu_usage"`
	SystemCpuUsage uint64 `json:"system_cpu_usage"`
}

func (processStats *LinuxProcessStats) getProcessTotalJiffies(prev, curr StatsSample) float64 {
	prevCPU := prev.stat.CPUStatAll
	currCPU := curr.stat.CPUStatAll

	previousTotal := (prevCPU.User + prevCPU.Nice + prevCPU.Idle + prevCPU.IOWait +
	 prevCPU.System + prevCPU.IRQ + prevCPU.SoftIRQ + prevCPU.Steal)
  currentTotal := (currCPU.User + currCPU.Nice + currCPU.Idle + currCPU.IOWait +
		currCPU.System + currCPU.IRQ + currCPU.SoftIRQ + currCPU.Steal)

	deltaTotal := currentTotal - previousTotal
	return float64(deltaTotal)
}

func (processStats *LinuxProcessStats) getProcessUsage(prev, curr StatsSample, index int) (uint64, uint64) {
	totalJiffies := processStats.getProcessTotalJiffies(prev, curr)
	userJiffies := curr.processes[index].Stat.Utime - prev.processes[index].Stat.Utime
	systemJiffies := curr.processes[index].Stat.Stime - prev.processes[index].Stat.Stime

  percentageUser := 100.0 * float64(userJiffies) / totalJiffies
	percentageSystem := 100.0 * float64(systemJiffies) / totalJiffies
  return uint64(math.Ceil(percentageUser)), uint64(math.Ceil(percentageSystem))
}

func (processStats *LinuxProcessStats) capToLong(number uint64) uint64 {
	if(number > math.MaxInt64) {
		return uint64(math.MaxInt64)
	}
  return number
}

func NewLinuxProcessesStats() []*LinuxProcessStats {
	processes := []*LinuxProcessStats{}

	previous, current := SharedStatsPeriod.GetStatsSamples()

	for i, _ := range current.processes {
		process := LinuxProcessStats{}
		process.CmdLine = current.processes[i].Cmdline
		process.Pid = current.processes[i].Status.Pid
		process.State = current.processes[i].Status.State
		process.MemVirtSize = current.processes[i].Statm.Size
		process.MemRssSize = current.processes[i].Statm.Resident
		process.MemLockSize = current.processes[i].Status.VmLck
		process.MemSwapSize = current.processes[i].Status.VmSwap
		process.Threads = current.processes[i].Status.Threads
		process.FDUsed = current.processes[i].Status.FDSize
		process.SignalsIgnored = process.capToLong(current.processes[i].Status.SigIgn - previous.processes[i].Status.SigIgn)
		process.SignalsCaught = process.capToLong(current.processes[i].Status.SigCgt - previous.processes[i].Status.SigCgt)
		process.UserCpuUsage, process.SystemCpuUsage = process.getProcessUsage(previous, current, i)
		process.VoluntaryContextSwitches = process.capToLong(current.processes[i].Status.VoluntaryCtxtSwitches - previous.processes[i].Status.VoluntaryCtxtSwitches)
		process.NonVoluntaryContextSwitches = process.capToLong(current.processes[i].Status.NonvoluntaryCtxtSwitches - previous.processes[i].Status.NonvoluntaryCtxtSwitches)
		process.IOReadBytes = current.processes[i].IO.ReadBytes - previous.processes[i].IO.ReadBytes
		process.IOWriteBytes = current.processes[i].IO.WriteBytes - previous.processes[i].IO.WriteBytes

		processes = append(processes, &process)
	}

	return processes
}
