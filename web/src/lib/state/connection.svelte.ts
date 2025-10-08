import { MessageType, type ICECandidateMessage, type Message } from '$lib/types/message';
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

	get ws() {
		return this.#ws;
	}

	// private event callbacks
	#onOpenCallback?: () => void;
	#onErrorCallback?: (error: unknown) => void;
	#onCloseCallback?: () => void;

	constructor(config: ConnectionStateConfig) {
		this.#baseUrl = config.baseUrl;
		this.#maxReconnectAttempts = config.maxReconnectAttempts ?? 3;

		this.#onOpenCallback = config.onOpen;
		this.#onErrorCallback = config.onError;
		this.#onCloseCallback = config.onClose;
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
		this.#pc = null;
		this.#dataChannel = null;

		if (this.#reconnectTimeout) {
			clearTimeout(this.#reconnectTimeout);
			this.#reconnectTimeout = null;
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

		this.#ws.onmessage = (event: MessageEvent) => {
			try {
				// assert Message type
				const data = JSON.parse(event.data) as Message;
				console.log('message received', data);
				if (typeof data.type !== 'string' || data.type.length === 0) {
					throw new Error('Invalid message format: missing type field');
				}
				switch (data.type) {
					case MessageType.Offer:
						break;
					case MessageType.Answer:
						break;
					case MessageType.ICECandidate:
						break;
					case MessageType.Error:
						console.error('Received error message:', data);
						break;
					case MessageType.WebRTCConnected:
						console.log('WebRTC connection established');
						break;
					case MessageType.RoomMeta:
						if ('roomId' in data) {
							this.roomId = data.roomId;
						}
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
					console.log('setting unknown thing');
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

	/**
	 * Initializes the process for p2p connection. Once it's done, the websocket will disconnect
	 * and the game from thereon will be handled in p2p.svelte.ts
	 */
	startGame(): void {}

	initiateP2P() {
		console.log('Initiating P2P connection');
		this.#pc = new RTCPeerConnection({
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

		this.#dataChannel = this.#pc.createDataChannel('game-channel', {
			ordered: true,
			maxRetransmits: 3
		});

		this.#dataChannel.onopen = () => {
			console.log('Data chanel opened. WebRTC is connected');
		};

		this.#dataChannel.onclose = () => {
			console.log('Data channel closed');
		};

		this.#dataChannel.onerror = () => {
			console.error('Data channel error');
		};

		this.#dataChannel.onmessage = (event) => {
			const data = JSON.parse(event.data);
			console.log('Data channel message received', data);
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

		this.#pc.onconnectionstatechange = () => {
			console.log('Peer connection state changed:', this.#pc?.connectionState);
			if (!this.#pc?.connectionState) {
				console.warn('No connection state');
				return;
			}
			switch (this.#pc.connectionState) {
				case 'connected':
					console.log('Peer connected!!!');
					break;
				case 'disconnected':
					console.warn('Peer disconnected');
					break;
				case 'failed':
					console.error('Peer connection failed');
					// handle
					break;
				case 'closed':
					console.log('Peer connection closed');
					break;
			}
		};

		this.#pc.oniceconnectionstatechange = () => {
			console.log("ICE Connection state changed:", this.#pc?.iceConnectionState);
		}

		try {}
	}
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
