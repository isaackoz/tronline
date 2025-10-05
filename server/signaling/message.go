package signaling

type Message struct {
	// "offer", "answer", "ice-candidate", "error", "webrtc-connected"
	Type string `json:"type"`
	// "host" or "client"
	Target string `json:"target,omitempty"`
	// "host" or "client"
	From string `json:"from,omitempty"`
	Data any    `json:"data,omitempty"`
}

// Specific message payloads
type OfferMessage struct {
	// "offer"
	Type   string `json:"type"`
	SDP    string `json:"sdp"`
	Target string `json:"target"`
}

type AnswerMessage struct {
	// "answer"
	Type   string `json:"type"`
	SDP    string `json:"sdp"`
	Target string `json:"target"`
}

type ICECandidateMessage struct {
	// "ice-candidate"
	Type      string       `json:"type"`
	Candidate ICECandidate `json:"candidate"`
	Target    string       `json:"target"`
}

type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
}

type ErrorMessage struct {
	Type    string `json:"type"` // "error"
	Message string `json:"message"`
}
