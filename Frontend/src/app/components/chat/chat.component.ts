import { Component, ElementRef, OnDestroy, OnInit, ViewChild } from "@angular/core";
import { Message } from "../../models/message.model";
import { WebsocketService } from "../../services/websocket.service";
import { AuthService } from "../../services/auth.service";
import { HttpClient, HttpHeaders } from "@angular/common/http";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";

@Component({
    selector: 'app-chat',
    templateUrl: './chat.component.html',
    styleUrl: './chat.component.scss',
    standalone: true,
    imports: [CommonModule, FormsModule]
})
export class ChatComponent implements OnInit, OnDestroy {
    @ViewChild('messagesContainer') private messagesContainer!: ElementRef;

    messages: Message[] = [];
    onlineUsers: string[] = [];
    newMessage: string = '';
    isLoadingMore: boolean = false;
    lastMessageId: number = 0;

    constructor(
        private websocketService: WebsocketService,
        public authService: AuthService,
        private http: HttpClient
    ) { }

    ngOnInit(): void {
        this.loadRecentMessages();

        this.websocketService.connect();

        this.websocketService.messages$.subscribe(message => {
            this.messages.push(message);
            setTimeout(() => this.scrollToBottom(), 0);
        });

        this.websocketService.onlineUsers$.subscribe(users => {
            this.onlineUsers = users;
        });
    }

    ngOnDestroy(): void {
        this.websocketService.disconnect();
    }

    loadRecentMessages(): void {
        const headers = new HttpHeaders({
            'Authorization': `Bearer ${this.authService.token}`
        });

        this.http.get<Message[]>(this.authService.API_URL + '/messages/recent', { headers })
            .subscribe({
                next: (messages) => {
                    if (messages != null) {
                        this.messages = messages;
                        if (messages.length > 0) {
                            this.lastMessageId = messages[0].id;
                        }
                        setTimeout(() => this.scrollToBottom(), 0);
                    }
                },
                error: (error) => {
                    console.error('Failed to load messages:', error);
                }
            });
    }

    loadPreviousMessages(): void {
        if (this.isLoadingMore || this.lastMessageId === 0) return;

        this.isLoadingMore = true;
        const headers = new HttpHeaders({
            'Authorization': `Bearer ${this.authService.token}`
        });

        this.http.get<Message[]>(`${this.authService.API_URL}/messages/before/${this.lastMessageId}`, { headers })
            .subscribe({
                next: (messages) => {
                    if (messages != null && messages.length > 0) {
                        this.messages = [...messages, ...this.messages];
                        this.lastMessageId = messages[0].id;
                    }
                    this.isLoadingMore = false;
                },
                error: (error) => {
                    console.error('Failed to load previous messages:', error);
                    this.isLoadingMore = false;
                }
            });
    }


    onScroll(event: any): void {
        const element = event.target;
        if (element.scrollTop <= element.clientHeight * 0.2) {
            this.loadPreviousMessages();
        }
    }

    scrollToBottom(): void {
        const element = this.messagesContainer.nativeElement;
        element.scrollTop = element.scrollHeight;
    }

    sendMessage(): void {
        if (this.newMessage.trim()) {
            this.websocketService.sendMessage(this.newMessage);
            this.newMessage = '';
        }
    }

    logout(): void {
        this.authService.logout().subscribe();
    }
}