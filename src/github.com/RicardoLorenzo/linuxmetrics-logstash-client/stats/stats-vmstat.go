package stats

type LinuxVMStats struct {
	PgFree uint64 `json:"pgfree"`
	PgpgIn uint64 `json:"pgpgin"`
	PgpgOut uint64 `json:"pgpgout"`
	PswpIn uint64 `json:"pswpin"`
	PswpOut uint64 `json:"pswpout"`
	PgFault uint64 `json:"pgfault"`
	PgMajFault uint64 `json:"pgmajfault"`
	NrMLock uint64 `json:"nr_mlock"`
	NrShMem uint64 `json:"nr_shmem"`
	NrDirty uint64 `json:"nr_dirty"`
	NrPageTablePages uint64 `json:"nr_page_table_pages"`
	NrSlab uint64 `json:"nr_slab"`
	NrMapped uint64 `json:"nr_mapped"`
	NrFreePages uint64 `json:"nr_free_pages"`
	NrAnonPages uint64 `json:"nr_anon_pages"`
}

func NewLinuxVMStats() *LinuxVMStats {
	vmStats := LinuxVMStats{}

	previous, current := SharedStatsPeriod.GetStatsSamples()

  // Additional memory access protection avoiding null references
  if SharedStatsPeriod.HasPreviousSamples() {
		vmStats.PgFree = current.vmstat.PageFree - previous.vmstat.PageFree
		vmStats.PgpgIn = current.vmstat.PagePagein - previous.vmstat.PagePagein
		vmStats.PgpgOut = current.vmstat.PagePageout - previous.vmstat.PagePageout
		vmStats.PswpIn = current.vmstat.PageSwapin - previous.vmstat.PageSwapin
		vmStats.PswpOut = current.vmstat.PageSwapout - previous.vmstat.PageSwapout
		vmStats.PgFault = current.vmstat.PageFault - previous.vmstat.PageFault
		vmStats.PgMajFault = current.vmstat.PageMajorFault - previous.vmstat.PageMajorFault
		vmStats.NrMLock = current.vmstat.NrMlock
		vmStats.NrShMem = current.vmstat.NrShmem
		vmStats.NrDirty = current.vmstat.NrDirty
		vmStats.NrPageTablePages = current.vmstat.NrPageTablePages
		vmStats.NrMapped = current.vmstat.NrMapped
		vmStats.NrFreePages = current.vmstat.NrFreePages
		vmStats.NrAnonPages = current.vmstat.NrAnonPages
  }

	return &vmStats
}
