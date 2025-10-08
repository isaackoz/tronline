import { MessageType, type Message } from '$lib/types/message';
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

	// All of the following are public reactive (runes) state
	roomId = $state<string | null>(null);
	reconnectAttempts = $state(0);

	connectionError = $state<string | null>(null);
	isConnected = $state(false);
	isConnecting = $state(true);
	isDisconnecting = $state(false);

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
		this.connectionError = null;
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
		this.isDisconnecting = true;
		this.#shouldReconnect = false;
		if (this.#reconnectTimeout) {
			clearTimeout(this.#reconnectTimeout);
			this.#reconnectTimeout = null;
		}

		if (this.#ws) {
			this.#ws.close();
			this.#ws = null;
		}
		this.isDisconnecting = false;
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
						this.disconnect();
						break;
					case MessageType.HostLeft:
						console.warn('Host has left the room');
						this.disconnect();
						break;
					case MessageType.GuestLeft:
						console.warn('Guest has left the room');
						break;
					case MessageType.GuestJoined:
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
