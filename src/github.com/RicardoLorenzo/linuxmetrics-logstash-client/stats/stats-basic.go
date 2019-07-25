package stats

import (
	"math"
)

type ProcessorStats struct {
	Cpu string `json:"cpu"`
  User uint64 `json:"user"`
	Nice uint64 `json:"nice"`
	System uint64 `json:"system"`
	IOWait uint64 `json:"iowait"`
	PercentageUtil uint64 `json:"percentageUtil"`
}

type LinuxBasicStats struct {
	Processors []*ProcessorStats `json:"processors"`
  AllProcessors *ProcessorStats `json:"allProcessors"`
	Processes uint64 `json:"processes"`
	ContextSwitches uint64 `json:"contextSwitches"`
	Interrupts uint64 `json:"interrupts"`
}

func (basicStats *LinuxBasicStats) getSingleCoreUsage(prev, curr StatsSample, index int) uint64 {
	/**
	 * https://rosettacode.org/wiki/Linux_CPU_utilization
	 */
  prevCPU := prev.stat.CPUStats[index]
	currCPU := curr.stat.CPUStats[index]

	previousTotal := (prevCPU.User + prevCPU.Nice + prevCPU.Idle + prevCPU.IOWait +
	 prevCPU.System + prevCPU.IRQ + prevCPU.SoftIRQ + prevCPU.Steal)
  currentTotal := (currCPU.User + currCPU.Nice + currCPU.Idle + currCPU.IOWait +
		currCPU.System + currCPU.IRQ + currCPU.SoftIRQ + currCPU.Steal)

  deltaIdle := currCPU.Idle - prevCPU.Idle
  deltaTotal := currentTotal - previousTotal

  percentage := (1.0 - float64(deltaIdle)/float64(deltaTotal)) * 100.0
  return uint64(math.Ceil(percentage))
}

func NewLinuxBasicStats() *LinuxBasicStats {
	basicStats := LinuxBasicStats{}

	previous, current := SharedStatsPeriod.GetStatsSamples()

  // Additional memory access protection avoiding null references
  if SharedStatsPeriod.HasPreviousSamples() {
		var allCPUPercentage uint64

    for i, _ := range current.stat.CPUStats {
			processorStat := new(ProcessorStats)
			cpuPercentage := basicStats.getSingleCoreUsage(previous, current, i)
      allCPUPercentage += cpuPercentage
			processorStat.Cpu = current.stat.CPUStats[i].Id
			processorStat.User = current.stat.CPUStats[i].User
			processorStat.Nice = current.stat.CPUStats[i].Nice
			processorStat.System = current.stat.CPUStats[i].System
			processorStat.IOWait = current.stat.CPUStats[i].IOWait
			processorStat.PercentageUtil = cpuPercentage

			basicStats.Processors = append(basicStats.Processors, processorStat)
    }

    processorStat := new(ProcessorStats)
		processorStat.User = current.stat.CPUStatAll.User
		processorStat.Nice = current.stat.CPUStatAll.Nice
		processorStat.System = current.stat.CPUStatAll.System
		processorStat.IOWait = current.stat.CPUStatAll.IOWait
		processorStat.PercentageUtil = allCPUPercentage / uint64(len(current.stat.CPUStats))

		basicStats.AllProcessors = processorStat

		basicStats.Processes = current.stat.Processes - previous.stat.Processes
		basicStats.ContextSwitches = current.stat.ContextSwitches - previous.stat.ContextSwitches
		basicStats.Interrupts = current.stat.Interrupts - previous.stat.Interrupts
  }

	return &basicStats
}
