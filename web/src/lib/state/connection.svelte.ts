import {
	MessageType,
	type AnswerMessage,
	type ICECandidate,
	type ICECandidateMessage,
	type Message,
	type OfferMessage,
	type WebRTCConnectedMessage
} from '$lib/types/message';
import { getContext, setContext } from 'svelte';

interface ConnectionStateConfig {
	/**
	 * The websocket URL to connect to.
	 * @example ws://localhost:4000/ws
	 */
	baseUrl: string;
	/**
	 * @default 3
	 */
	maxReconnectAttempts?: number;

	// callbacks for connection events
	onOpen?: () => void;
	onError?: (error: unknown) => void;
	onClose?: () => void;
	onWebRTCReady?: () => void;
	onGameMessage?: (data: unknown) => void;
}

export class ConnectionState {
	// private state
	#ws: WebSocket | null = $state(null);
	#baseUrl: string;
	#maxReconnectAttempts: number;
	#shouldReconnect: boolean = false;
	#reconnectInterval = 2000;
	#reconnectTimeout: number | null = null;
	#role: 'host' | 'client' | null = null;
	#pc: RTCPeerConnection | null = null;
	#dataChannel: RTCDataChannel | null = null;

	// All of the following are public reactive (runes) state
	roomId = $state<string | null>(null);
	reconnectAttempts = $state(0);

	connectionError = $state<string | null>(null);
	roomError = $state<string | null>(null);
	isConnected = $state(false);
	isConnecting = $state(false);
	isGuestConnected = $state(false);
	isWebRTCConnected = $state(false);
	isP2PConnecting = $state(false);

	// private event callbacks
	#onOpenCallback?: () => void;
	#onErrorCallback?: (error: unknown) => void;
	#onCloseCallback?: () => void;
	#onWebRTCReadyCallback?: () => void;
	#onGameMessageCallback?: (data: unknown) => void;

	constructor(config: ConnectionStateConfig) {
		this.#baseUrl = config.baseUrl;
		this.#maxReconnectAttempts = config.maxReconnectAttempts ?? 3;

		this.#onOpenCallback = config.onOpen;
		this.#onErrorCallback = config.onError;
		this.#onCloseCallback = config.onClose;
		this.#onWebRTCReadyCallback = config.onWebRTCReady;
		this.#onGameMessageCallback = config.onGameMessage;
	}

	connect(role: 'client' | 'host', roomId?: string): void {
		if (this.#ws && this.#ws.readyState === WebSocket.OPEN) {
			console.warn('WebSocket is already connected');
			return;
		}
		this.#pc = null;
		this.#dataChannel = null;
		this.connectionError = null;
		this.isGuestConnected = false;
		this.isWebRTCConnected = false;
		this.roomError = null;
		this.isConnecting = true;
		this.roomId = roomId ?? null;
		this.#role = role;
		this.#shouldReconnect = false;

		try {
			let url = this.#baseUrl;
			if (role === 'client') {
				if (!roomId) throw new Error('Room ID is required to join a game');
				url += `?role=client&roomId=${roomId}`;
			} else {
				url += `?role=host`;
			}
			this.#ws = new WebSocket(url);
			this.#setupEventListeners();
		} catch (err) {
			console.error('Failed to create WebSocket', err);
			this.#handleReconnect();
		} finally {
			this.isConnecting = false;
		}
	}

	disconnect(): void {
		this.#shouldReconnect = false;
		this.connectionError = null;
		this.isGuestConnected = false;
		this.isWebRTCConnected = false;
		this.isP2PConnecting = false;

		if (this.#reconnectTimeout) {
			clearTimeout(this.#reconnectTimeout);
			this.#reconnectTimeout = null;
		}
		if (this.#dataChannel) {
			this.#dataChannel.close();
			this.#dataChannel = null;
		}
		if (this.#pc) {
			this.#pc.close();
			this.#pc = null;
		}
		if (this.#ws) {
			this.#ws.close(1000);
			this.#ws = null;
		}
	}

	#setupEventListeners() {
		if (!this.#ws) return;

		this.#ws.onopen = () => {
			this.isConnected = true;
			this.reconnectAttempts = 0;
			// only attempt to reconnect after a successful connection
			this.#shouldReconnect = true;
			this.#onOpenCallback?.();
		};

		this.#ws.onmessage = async (event: MessageEvent) => {
			try {
				// assert Message type
				const data = JSON.parse(event.data) as Message;
				console.log('message received', data);
				if (typeof data.type !== 'string' || data.type.length === 0) {
					throw new Error('Invalid message format: missing type field');
				}
				switch (data.type) {
					case MessageType.Offer:
						await this.handleOffer(data);
						break;
					case MessageType.Answer:
						await this.handleAnswer(data);
						break;
					case MessageType.ICECandidate:
						await this.handleIceCandidate(data.candidate);
						break;
					case MessageType.Error:
						console.error('Received error message:', data);
						break;
					case MessageType.WebRTCConnected:
						console.log('WebRTC connection established');
						break;
					case MessageType.RoomMeta:
						this.roomId = data.roomId;
						break;
					case MessageType.RoomClosed:
						console.warn('Room has been closed by the host');
						this.roomError = 'The room has been closed by the host.';
						this.disconnect();
						break;
					case MessageType.HostLeft:
						this.roomError = 'The host has left the room.';
						this.disconnect();
						break;
					case MessageType.GuestLeft:
						this.isGuestConnected = false;
						console.warn('Guest has left the room');
						break;
					case MessageType.GuestJoined:
						this.isGuestConnected = true;
						console.log('A guest has joined the room');
						break;
					default:
						console.warn('Unknown message type');
				}
			} catch (err) {
				console.error('Failed to parse message', err);
			}
		};

		this.#ws.onerror = (error: Event) => {
			console.error('WebSocket error', error);
			this.isConnected = false;
			this.connectionError = 'Unknown connection error.';
			this.#onErrorCallback?.(error);
		};

		this.#ws.onclose = (closeEvent: CloseEvent) => {
			console.log('WebSocket closed');
			this.isConnected = false;
			this.#onCloseCallback?.();
			switch (closeEvent.code) {
				case 3001:
				case 3002:
				case 3003:
				case 3004:
				case 3005:
					this.#shouldReconnect = false;
					this.connectionError = closeEvent.reason;
					break;
				default:
					this.connectionError = 'An unknown connection error occurred.';
			}
			if (this.#shouldReconnect) {
				console.log('Attempting to reconnect...');
				this.#handleReconnect();
			}
		};
	}

	#handleReconnect(): void {
		if (this.reconnectAttempts >= this.#maxReconnectAttempts) {
			console.error('Max reconnect attempts reached');
			return;
		}

		this.reconnectAttempts++;

		console.log(`Reconnect attempt ${this.reconnectAttempts}`);

		this.#reconnectTimeout = window.setTimeout(() => {
			if (!this.#role) {
				console.error('Role is not set. Cannot reconnect.');
				return;
			}
			this.connect(this.#role, this.roomId ?? undefined);
		}, this.#reconnectInterval);
	}

	#onWebRTCConnectionReady() {
		console.log('WebRTC connection is ready');
		this.isWebRTCConnected = true;

		// notify server that webrtc is connected
		if (this.#ws && this.#ws.readyState === WebSocket.OPEN) {
			const msg: WebRTCConnectedMessage = {
				type: MessageType.WebRTCConnected
			};
			this.#ws.send(JSON.stringify(msg));
		}

		// close the websocket after WebRTC is ready
		if (this.#ws && this.#ws.readyState === WebSocket.OPEN) {
			this.#shouldReconnect = false;
			this.#ws.close(1000, 'WebRTC Connected');
			this.#ws = null;
			console.log('WebSocket closed after WebRTC connection established');
		}
		this.#onWebRTCReadyCallback?.();
	}

	// Host calls this
	async initiateP2P(): Promise<void> {
		try {
			this.isP2PConnecting = true;
			console.log('Initiating P2P connection');
			console.log('this:', this);
			console.log('this.constructor.name:', this.constructor.name);
			console.log('Is ConnectionState?', this instanceof ConnectionState);
			if (!this.#ws) {
				throw new Error('WebSocket is not connected');
			}
			this.#pc = createRtcPeerConnection();

			this.#dataChannel = this.#pc.createDataChannel('game-channel', {
				ordered: true,
				maxRetransmits: 3
			});

			this.#setupDataChannel();

			this.#pc.onicecandidate = (event) => {
				if (event.candidate) {
					console.log('New ICE candidate', event.candidate);
					if (
						!event.candidate.candidate ||
						!event.candidate.sdpMid ||
						event.candidate.sdpMLineIndex === null
					) {
						console.error('Invalid ICE candidate');
						return;
					}
					const candidateMessage: ICECandidateMessage = {
						type: MessageType.ICECandidate,
						target: 'client', // should be set on server too
						//from is set on server,
						candidate: {
							candidate: event.candidate.candidate,
							sdpMid: event.candidate.sdpMid,
							sdpMLineIndex: event.candidate.sdpMLineIndex
						}
					};
					this.#ws?.send(JSON.stringify(candidateMessage));
				}
			};

			this.#setupConnectionStateHandlers();

			// create and send an offer (this message gets proxied through the ws server)
			try {
				const offer = await this.#pc.createOffer();
				if (!offer.sdp) {
					throw new Error('Failed to create offer');
				}
				await this.#pc.setLocalDescription(offer);
				const offerMsg: OfferMessage = {
					type: MessageType.Offer,
					sdp: offer.sdp,
					// the server will manage the target/from depending on the role in the room
					target: 'client' // should be set on server too
					// from: "host"
				};
				this.#ws.send(JSON.stringify(offerMsg));
			} catch (err) {
				console.error('Failed to create/send offer', err);
			}
		} catch (err) {
			console.error('Failed to initiate P2P', err);
			this.connectionError = 'Failed to initiate P2P connection.';
		} finally {
			this.isP2PConnecting = false;
		}
	}

	// Guest handles the offer
	async handleOffer(offerData: OfferMessage): Promise<void> {
		console.log('handling guest offer', offerData);
		if (!this.#ws) {
			throw new Error('WebSocket is not connected');
		}

		this.#pc = createRtcPeerConnection();

		this.#pc.ondatachannel = (event) => {
			console.log('Data channel received', event.channel);
			this.#dataChannel = event.channel;
			this.#setupDataChannel();
		};

		this.#pc.onicecandidate = (event) => {
			if (event.candidate) {
				console.log('New ICE candidate', event.candidate);
				if (
					!event.candidate.candidate ||
					!event.candidate.sdpMid ||
					event.candidate.sdpMLineIndex === null
				) {
					console.error('Invalid ICE candidate');
					return;
				}
				const candidateMessage: ICECandidateMessage = {
					type: MessageType.ICECandidate,
					target: 'host',
					candidate: {
						candidate: event.candidate.candidate,
						sdpMid: event.candidate.sdpMid,
						sdpMLineIndex: event.candidate.sdpMLineIndex
					}
				};
				this.#ws?.send(JSON.stringify(candidateMessage));
			}
		};

		this.#setupConnectionStateHandlers();

		try {
			await this.#pc.setRemoteDescription(
				new RTCSessionDescription({
					type: 'offer',
					sdp: offerData.sdp
				})
			);
			console.log('Remote description set with offer');
		} catch (err) {
			console.error('Failed to set remote description', err);
			return;
		}

		try {
			const answer = await this.#pc.createAnswer();
			if (!answer.sdp) {
				throw new Error('Failed to create answer');
			}
			await this.#pc.setLocalDescription(answer);

			console.log('Local description set with answer');

			const answerMsg: AnswerMessage = {
				type: MessageType.Answer,
				sdp: answer.sdp,
				target: 'host'
			};
			this.#ws.send(JSON.stringify(answerMsg));
		} catch (err) {
			console.error('Failed to create/send answer', err);
		}
	}

	async handleAnswer(answerData: AnswerMessage) {
		console.log('Handling answer', answerData);
		if (!this.#pc) {
			throw new Error('PeerConnection is not initialized');
		}
		try {
			await this.#pc.setRemoteDescription(
				new RTCSessionDescription({
					type: 'answer',
					sdp: answerData.sdp
				})
			);
			console.log('Remote description set with answer');
		} catch (err) {
			console.error('Failed to handle answer', err);
		}
	}

	async handleIceCandidate(candidateData: ICECandidate) {
		console.log('Handling ICE candidate', candidateData);
		if (!candidateData.candidate) return;
		if (!this.#pc) {
			throw new Error('PeerConnection is not initialized');
		}
		try {
			await this.#pc.addIceCandidate(
				new RTCIceCandidate({
					candidate: candidateData.candidate,
					sdpMid: candidateData.sdpMid,
					sdpMLineIndex: candidateData.sdpMLineIndex
				})
			);

			console.log('Ice candidate added');
		} catch (err) {
			console.error('Failed to add ICE candidate', err);
		}
	}

	#setupDataChannel() {
		if (!this.#dataChannel) return;

		this.#dataChannel.onopen = () => {
			console.log('Data chanel opened. WebRTC is connected');
			this.#onWebRTCConnectionReady();
		};

		this.#dataChannel.onclose = () => {
			console.log('Data channel closed');
			this.isWebRTCConnected = false;
		};

		this.#dataChannel.onerror = () => {
			console.error('Data channel error');
		};

		this.#dataChannel.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				console.log('Game message received', data);
				this.#onGameMessageCallback?.(data);
			} catch (err) {
				console.error('Failed to parse game message', err);
			}
		};
	}

	#setupConnectionStateHandlers() {
		if (!this.#pc) return;

		this.#pc.onconnectionstatechange = () => {
			console.log('Peer connection state changed:', this.#pc?.connectionState);
			if (!this.#pc?.connectionState) return;

			switch (this.#pc.connectionState) {
				case 'connected':
					console.log('Peer connected!!!');
					break;
				case 'disconnected':
					console.warn('Peer disconnected');
					this.isWebRTCConnected = false;
					break;
				case 'failed':
					console.error('Peer connection failed');
					this.isWebRTCConnected = false;
					break;
				case 'closed':
					console.log('Peer connection closed');
					this.isWebRTCConnected = false;
					break;
			}
		};

		this.#pc.oniceconnectionstatechange = () => {
			console.log('ICE Connection state changed:', this.#pc?.iceConnectionState);
		};
	}

	sendGameMessage(data: unknown): boolean {
		if (!this.#dataChannel || this.#dataChannel.readyState !== 'open') {
			console.warn('Data channel is not open; cant send message');
			return false;
		}

		try {
			this.#dataChannel.send(JSON.stringify(data));
			return true;
		} catch (err) {
			console.error('Failed to send game message', err);
			return false;
		}
	}

	cleanup() {
		if (this.#dataChannel) {
			this.#dataChannel.close();
			this.#dataChannel = null;
		}
		if (this.#pc) {
			this.#pc.close();
			this.#pc = null;
		}
		this.disconnect();
	}
}

/**
 * Returns a new RTCPeerConnection with pre-configured STUN/TURN servers.
 */
function createRtcPeerConnection() {
	return new RTCPeerConnection({
		iceServers: [
			{
				urls: [
					'stun:stun.cloudflare.com:3478',
					'stun:stun.cloudflare.com:53',
					'turn:turn.cloudflare.com:3478?transport=udp',
					'turn:turn.cloudflare.com:53?transport=udp',
					'turn:turn.cloudflare.com:3478?transport=tcp',
					'turn:turn.cloudflare.com:80?transport=tcp',
					'turns:turn.cloudflare.com:5349?transport=tcp',
					'turns:turn.cloudflare.com:443?transport=tcp'
				]
			}
		]
	});
}

// unique key for the connection state context
const CONNECTION_STATE_KEY = Symbol('CONNECTIONSTATE');

// helper to set connection state
export function setConnectionState(initialProps: ConnectionStateConfig) {
	return setContext(CONNECTION_STATE_KEY, new ConnectionState(initialProps));
}

// helper to get the connection state
export function getConnectionState() {
	return getContext<ReturnType<typeof setConnectionState>>(CONNECTION_STATE_KEY);
}
