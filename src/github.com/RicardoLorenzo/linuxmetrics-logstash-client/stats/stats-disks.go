package stats

import (
	"math"
)

/*
 * https://lkml.org/lkml/2015/8/17/269
 *
 * It's cast in stone. There are too many places all over the kernel,
 * especially in a huge number of file systems, which assume that the
 * sector size is 512 bytes. So above the block layer, the sector size
 * is always going to be 512.
 *
 * This is actually better for user space programs using /proc/diskstats,
 * since they don't need to know whether a particular underlying hardware
 * is using 512, 4k, (or if the HDD manufacturers fantasies become true
 * 32k or 64k) sector sizes.
 *
 * For similar reason, st_blocks in struct size is always in units of
 * 512 bytes. We don't want to force userspace to have to figure out
 * whether the underlying file system is using 1k, 2k, or 4k. For that
 * reason the units of st_blocks is always going to be 512 bytes, and
 * this is hard-coded in the POSIX standard.
 */
const sectorSize uint64 = 512

const MEGABYTE float64 = 1024 * 1024

/**
 * All fields except field 9 are cumulative since boot.  Field 9 should
 * go to zero as I/Os complete; all others only increase (unless they
 * overflow and wrap).  Yes, these are (32-bit or 64-bit) unsigned long
 * (native word size) numbers, and on a very busy or long-lived system they
 * may wrap.
 *
 * https://www.kernel.org/doc/Documentation/iostats.txt
 */
type LinuxDiskStats struct {
	Name string `json:"name"`
	ReadIOs uint64 `json:"read_io"`
	WriteIOs uint64 `json:"write_io"`
	ReadMerges uint64 `json:"read_io_merged"`
	WriteMerges uint64 `json:"write_io_merged"`
	IOTicks uint64 `json:"io_ticks"`
	QueueSize uint64 `json:"queue_size"`
	TimeInQueue uint64 `json:"time_in_queue"`
	ReadMBPerSecond uint64 `json:"read_mbps"`
	WriteMBPerSecond uint64 `json:"write_mbps"`
}

func (disksStats *LinuxDiskStats) getMBPerSecond(sectors uint64) uint64 {
	return uint64(math.Floor(float64(sectors * sectorSize) / MEGABYTE))
}

func NewLinuxDisksStats() []*LinuxDiskStats {
	disksStats := []*LinuxDiskStats{}

	previous, current := SharedStatsPeriod.GetStatsSamples()

  // TODO In a very busy systems this counters may wrap. This is not curently expected
	for i, _ := range current.diskstats {
		diskStats := LinuxDiskStats{}
		diskStats.Name = current.diskstats[i].Name
		diskStats.ReadIOs = current.diskstats[i].ReadIOs - previous.diskstats[i].ReadIOs
		diskStats.WriteIOs = current.diskstats[i].WriteIOs - previous.diskstats[i].WriteIOs
		diskStats.ReadMerges = current.diskstats[i].ReadMerges - previous.diskstats[i].ReadMerges
		diskStats.WriteMerges = current.diskstats[i].WriteMerges - previous.diskstats[i].WriteMerges
		diskStats.QueueSize = current.diskstats[i].InFlight
		diskStats.TimeInQueue = current.diskstats[i].TimeInQueue - previous.diskstats[i].TimeInQueue
		diskStats.ReadMBPerSecond = diskStats.getMBPerSecond(current.diskstats[i].ReadSectors - previous.diskstats[i].ReadSectors)
		diskStats.WriteMBPerSecond = diskStats.getMBPerSecond(current.diskstats[i].WriteSectors - previous.diskstats[i].WriteSectors)
		disksStats = append(disksStats, &diskStats)
	}

	return disksStats
}
