package udptracker

type udpAnnounceResponse struct {
  udpMessageHeader
  Interval int32
  Leechers int32
  Seeders int32
}

func (h *udpMessageHeader) SetTransactionID(id int32) {
  h.TransactionID = id
}
