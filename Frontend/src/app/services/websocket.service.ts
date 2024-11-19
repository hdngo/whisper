import { Injectable } from "@angular/core";
import { BehaviorSubject, filter, Observable } from "rxjs";
import { Message } from "../models/message.model";
import { AuthService } from "./auth.service";

@Injectable({
    providedIn: 'root'
})
export class WebsocketService {
    private socket: WebSocket | null = null;
    private readonly WS_URL = `ws://${window.location.host}/api/ws`;

    private onlineUsersSubject = new BehaviorSubject<string[]>([]);
    private messagesSubject = new BehaviorSubject<Message | null>(null);

    constructor(private authService: AuthService) { }

    public connect(): void {
        if (this.socket) {
            this.socket.close();
        }

        // Add token as a subprotocol
        const token = this.authService.token;
        if (!token) {
            console.error('No token available');
            return;
        }

        this.socket = new WebSocket(this.WS_URL, [`access_token|${token}`]);

        this.socket.onmessage = (event) => {
            try {
                const wsMessage = JSON.parse(event.data);

                switch (wsMessage.type) {
                    case 'chat':
                        this.messagesSubject.next(wsMessage.payload);
                        break;

                    case 'users':
                        this.onlineUsersSubject.next(wsMessage.payload);
                        break;

                    case 'join':
                    case 'leave':
                        // Handle join/leave notifications if needed
                        break;
                }
            } catch (e) {
                console.error('Error parsing message:', e);
            }
        };
    }

    public disconnect(): void {
        if (this.socket) {
            this.socket.close();
            this.socket = null;
        }
    }

    public sendMessage(content: string): void {
        if (this.socket?.readyState === WebSocket.OPEN) {
            this.socket.send(content);
        }
    }

    public get onlineUsers$(): Observable<string[]> {
        return this.onlineUsersSubject.asObservable();
    }

    public get messages$(): Observable<Message> {
        return this.messagesSubject.asObservable().pipe(
            filter((message): message is Message => message !== null)
        );
    }
}