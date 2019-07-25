package stats

/**
 * See MIB-II reference
 * https://tools.ietf.org/html/rfc1213
 */
type LinuxNetworkStats struct {
	// IP
	IpForwarding uint64 `json:"ip_forwarding"`
	IpForwDatagrams uint64 `json:"ip_forwarded"`
  IpInReceives uint64 `json:"ip_in_received"`
	IpInHdrErrors uint64 `json:"ip_in_header_errors"`
	IpInAddrErrors uint64 `json:"ip_in_addr_errors"`
	IpInDiscards  uint64 `json:"ip_in_discarded"`
	// Discarded because of an unknown or unsupported protocol
	IpInUnknownProtos uint64 `json:"ip_in_unknown"`
	IpInDelivers uint64 `json:"ip_in_delivered"`
  // Does not include any of the IpForwDatagrams
	IpOutRequests uint64 `json:"ip_out_requests"`
	IpOutNoRoutes uint64 `json:"ip_out_noroute"`
	IpOutDiscards uint64 `json:"ip_out_discarded"`
	// TCP
	TcpRtoMax uint64 `json:"tcp_rto_max"`
	TcpMaxConn uint64 `json:"tcp_max_connections"`
	/**
	 * The number of times TCP connections have made a
   * direct transition to the SYN-SENT state from the
   * CLOSED state.
	 */
	TcpActiveOpens uint64 `json:"tcp_active_opened"`
	/**
	 * The number of times TCP connections have made a
   * direct transition to the SYN-RCVD state from the
   * LISTEN state.
	 */
	TcpPassiveOpens uint64 `json:"tcp_active_opened"`
	/**
	 * TCP connections for which the current state is
	 * either ESTABLISHED or CLOSE-WAIT
	 */
	TcpCurrEstab uint64 `json:"tcp_current_established"`
	TcpEstabResets uint64 `json:"tcp_established_reset"`
	TcpRetransSegs uint64 `json:"tcp_retransmited_seg"`
	TcpInSegs uint64 `json:"tcp_in_seg"`
	TcpOutSegs uint64 `json:"tcp_out_seg"`
	TcpInErrs  uint64 `json:"tcp_in_error"`
	TcpOutRsts  uint64 `json:"tcp_out_rst"`
	/**
	 * I'm summarizing in a single field all the receive and transmit
	 * queues from all sockets.
	 *
	 * https://www.kernel.org/doc/Documentation/networking/proc_net_tcp.txt
	 */
	TotalTCPSockets uint64 `json:"total_tcp_sockets"`
	TotalTCPRxQueue uint64 `json:"total_tcp_rx_queue"`
	TotalTCPTxQueue uint64 `json:"total_tcp_tx_queue"`
}

func NewLinuxNetworkStats() *LinuxNetworkStats {
	networkStats := LinuxNetworkStats{}

  var rxQueue uint64 = 0
	var txQueue uint64 = 0
	previous, current := SharedStatsPeriod.GetStatsSamples()

  networkStats.IpForwarding = current.snmp.IpForwarding
	networkStats.IpForwDatagrams = current.snmp.IpForwDatagrams - previous.snmp.IpForwDatagrams
	networkStats.IpInReceives = current.snmp.IpInReceives - previous.snmp.IpInReceives
	networkStats.IpInHdrErrors = current.snmp.IpInHdrErrors - previous.snmp.IpInHdrErrors
	networkStats.IpInAddrErrors = current.snmp.IpInAddrErrors - previous.snmp.IpInAddrErrors
	networkStats.IpInDiscards = current.snmp.IpInDiscards - previous.snmp.IpInDiscards
	networkStats.IpInUnknownProtos = current.snmp.IpInUnknownProtos - previous.snmp.IpInUnknownProtos
	networkStats.IpInDelivers = current.snmp.IpInDelivers - previous.snmp.IpInDelivers
	networkStats.IpOutRequests = current.snmp.IpOutRequests - previous.snmp.IpOutRequests
	networkStats.IpOutNoRoutes = current.snmp.IpOutNoRoutes - previous.snmp.IpOutNoRoutes
	networkStats.IpOutDiscards = current.snmp.IpOutDiscards - previous.snmp.IpOutDiscards
	networkStats.TcpRtoMax = current.snmp.TcpRtoMax
	networkStats.TcpMaxConn = current.snmp.TcpMaxConn
	networkStats.TcpActiveOpens = current.snmp.TcpActiveOpens - previous.snmp.TcpActiveOpens
	networkStats.TcpPassiveOpens = current.snmp.TcpPassiveOpens - previous.snmp.TcpPassiveOpens
	networkStats.TcpCurrEstab = current.snmp.TcpCurrEstab
	networkStats.TcpEstabResets = current.snmp.TcpEstabResets - previous.snmp.TcpEstabResets
	networkStats.TcpRetransSegs = current.snmp.TcpRetransSegs - previous.snmp.TcpRetransSegs
	networkStats.TcpInSegs = current.snmp.TcpInSegs - previous.snmp.TcpInSegs
	networkStats.TcpOutSegs = current.snmp.TcpOutSegs - previous.snmp.TcpOutSegs
	networkStats.TcpInErrs = current.snmp.TcpInErrs - previous.snmp.TcpInErrs
	networkStats.TcpOutRsts = current.snmp.TcpOutRsts - previous.snmp.TcpOutRsts

  // TX and RX queues aggregation
	for i, _ := range current.tcpsockets {
		rxQueue += current.tcpsockets[i].NetSocket.RxQueue
		txQueue += current.tcpsockets[i].NetSocket.TxQueue
	}

  networkStats.TotalTCPSockets = uint64(len(current.tcpsockets))
	networkStats.TotalTCPRxQueue = rxQueue
	networkStats.TotalTCPTxQueue = txQueue

	return &networkStats
}
