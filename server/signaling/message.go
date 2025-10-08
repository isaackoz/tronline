package signaling

type MessageType string

const (
	MessageTypeOffer           MessageType = "offer"
	MessageTypeAnswer          MessageType = "answer"
	MessageTypeICECandidate    MessageType = "ice-candidate"
	MessageTypeError           MessageType = "error"
	MessageTypeWebRTCConnected MessageType = "webrtc-connected"

	MessageTypeRoomMeta MessageType = "room-meta"

	MessageEventTypeHostLeft    MessageType = "host-left"
	MessageEventTypeGuestLeft   MessageType = "guest-left"
	MessageEventTypeGuestJoined MessageType = "guest-joined"
	MessageEventRoomClosed      MessageType = "room-closed"
)

// Message interface - all messages must implement GetType()
type Message interface {
	GetType() MessageType
}

type EventMessage struct {
	Type     MessageType `json:"type"`
	Metadata any         `json:"metadata,omitempty"`
}

func (m EventMessage) GetType() MessageType {
	return m.Type
}

type RoomMetaMessage struct {
	Type   MessageType `json:"type"`
	RoomId string      `json:"roomId"`
}

func (m RoomMetaMessage) GetType() MessageType {
	return m.Type
}

// OfferMessage represents a WebRTC offer
type OfferMessage struct {
	Type   MessageType `json:"type"`
	SDP    string      `json:"sdp"`
	Target string      `json:"target"`
	From   string      `json:"from,omitempty"`
}

func (m OfferMessage) GetType() MessageType {
	return MessageTypeOffer
}

// AnswerMessage represents a WebRTC answer
type AnswerMessage struct {
	Type   MessageType `json:"type"`
	SDP    string      `json:"sdp"`
	Target string      `json:"target"`
	From   string      `json:"from,omitempty"`
}

func (m AnswerMessage) GetType() MessageType {
	return MessageTypeAnswer
}

// ICECandidateMessage represents an ICE candidate exchange
type ICECandidateMessage struct {
	Type      MessageType  `json:"type"`
	Candidate ICECandidate `json:"candidate"`
	Target    string       `json:"target"`
	From      string       `json:"from,omitempty"`
}

func (m ICECandidateMessage) GetType() MessageType {
	return MessageTypeICECandidate
}

// ICECandidate represents a WebRTC ICE candidate
type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
}

// ErrorMessage represents an error condition
type ErrorMessage struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
	Target  string      `json:"target,omitempty"`
	From    string      `json:"from,omitempty"`
}

func (m ErrorMessage) GetType() MessageType {
	return MessageTypeError
}

// WebRTCConnectedMessage indicates successful WebRTC connection
type WebRTCConnectedMessage struct {
	Type   MessageType `json:"type"`
	Target string      `json:"target,omitempty"`
	From   string      `json:"from,omitempty"`
	Data   any         `json:"data,omitempty"`
}

func (m WebRTCConnectedMessage) GetType() MessageType {
	return MessageTypeWebRTCConnected
}
