package udptracker

type action int32

// udp tracker action
const (
	actionConnect  action = 0
	actionAnnounce action = 1
	actionError    action = 3
)
