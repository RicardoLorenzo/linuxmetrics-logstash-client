package stats

import (
	"encoding/json"
	"log"
)

type JSONStats struct {
	Type string  `json:"type"`
	Hostname string  `json:"hostname"`
	BasicStats *LinuxBasicStats `json:"basic"`
	Vmstat *LinuxVMStats `json:"vmstat"`
	NetworkStats *LinuxNetworkStats `json:"network"`
	Processes []*LinuxProcessStats `json:"processes"`
	Disks []*LinuxDiskStats `json:"disks"`
}

func NewJSONStats() *JSONStats {
  stats := JSONStats{}
	stats.Type = "osmetrics"
	return &stats
}

func (jsonstats *JSONStats) GetStats() (string, error) {
	SharedStatsPeriod.WaitForSamples()

  _, current := SharedStatsPeriod.GetStatsSamples()
	jsonstats.Hostname = current.hostname

  jsonstats.BasicStats = NewLinuxBasicStats()
	jsonstats.Vmstat = NewLinuxVMStats()
	jsonstats.NetworkStats = NewLinuxNetworkStats()
	jsonstats.Processes = NewLinuxProcessesStats()
	jsonstats.Disks = NewLinuxDisksStats()

	response, err := json.Marshal(jsonstats)
	if err != nil {
    log.Fatalln("Fail to encode JSON stats", err)
  }

  return string(response), nil
}
