// generated from /server/signaling/message.go using ai
// prompt: convert these go types to typescript types. make it a discriminated union type.

// MessageType enum
export enum MessageType {
	Offer = 'offer',
	Answer = 'answer',
	ICECandidate = 'ice-candidate',
	Error = 'error',
	WebRTCConnected = 'webrtc-connected',
	RoomMeta = 'room-meta',
	HostLeft = 'host-left',
	GuestLeft = 'guest-left',
	GuestJoined = 'guest-joined',
	RoomClosed = 'room-closed'
}

// ICE Candidate type
export interface ICECandidate {
	candidate: string;
	sdpMid: string;
	sdpMLineIndex: number;
}

// Event messages (host-left, guest-left, guest-joined, room-closed)
export interface EventMessage {
	type:
		| MessageType.HostLeft
		| MessageType.GuestLeft
		| MessageType.GuestJoined
		| MessageType.RoomClosed;
	metadata?: unknown;
}

// Room metadata message
export interface RoomMetaMessage {
	type: MessageType.RoomMeta;
	roomId: string;
}

// WebRTC Offer message
export interface OfferMessage {
	type: MessageType.Offer;
	sdp: string;
	target: string;
	from?: string;
}

// WebRTC Answer message
export interface AnswerMessage {
	type: MessageType.Answer;
	sdp: string;
	target: string;
	from?: string;
}

// ICE Candidate message
export interface ICECandidateMessage {
	type: MessageType.ICECandidate;
	candidate: ICECandidate;
	target: string;
	from?: string;
}

// Error message
export interface ErrorMessage {
	type: MessageType.Error;
	message: string;
	target?: string;
	from?: string;
}

// WebRTC Connected message
export interface WebRTCConnectedMessage {
	type: MessageType.WebRTCConnected;
}

// Discriminated union of all message types
export type Message =
	| EventMessage
	| RoomMetaMessage
	| OfferMessage
	| AnswerMessage
	| ICECandidateMessage
	| ErrorMessage
	| WebRTCConnectedMessage;
