import { Injectable } from "@angular/core";
import { BehaviorSubject, map, Observable } from "rxjs";
import { AuthResponse, LoginRequest, RegisterRequest } from "../models/auth.model";
import { HttpClient } from "@angular/common/http";

@Injectable({
    providedIn: 'root'
})
export class AuthService {
    public readonly API_URL = '/api';
    private currentUserSubject: BehaviorSubject<string | null>;
    private tokenSubject: BehaviorSubject<string | null>;

    constructor(private http: HttpClient) {
        this.currentUserSubject = new BehaviorSubject<string | null>(
            localStorage.getItem('currentUser')
        );
        this.tokenSubject = new BehaviorSubject<string | null>(
            localStorage.getItem('token')
        );
    }

    public get currentuser(): string | null {
        return this.currentUserSubject.value;
    }

    public get token(): string | null {
        return this.tokenSubject.value;
    }

    public get isLoggedIn(): boolean {
        return !!this.currentUserSubject.value && !!this.tokenSubject.value;
    }

    public clearState(): void {
        localStorage.removeItem('currentUser');
        localStorage.removeItem('token');
        this.currentUserSubject.next(null);
        this.tokenSubject.next(null);
    }

    public login(credentials: LoginRequest): Observable<AuthResponse> {
        return this.http.post<AuthResponse>(`${this.API_URL}/auth/login`, credentials)
            .pipe(
                map(response => {
                    localStorage.setItem('currentUser', response.username);
                    localStorage.setItem('token', response.token);
                    this.currentUserSubject.next(response.username);
                    this.tokenSubject.next(response.token);
                    return response;
                })
            );
    }

    public register(credentials: RegisterRequest): Observable<AuthResponse> {
        return this.http.post<AuthResponse>(`${this.API_URL}/auth/register`, credentials)
            .pipe(
                map(response => {
                    localStorage.setItem('currentUser', response.username);
                    localStorage.setItem('token', response.token);
                    this.currentUserSubject.next(response.username);
                    this.tokenSubject.next(response.token);
                    return response;
                })
            );
    }

    public logout(): Observable<void> {
        return this.http.post<void>(`${this.API_URL}/auth/logout`, null)
            .pipe(
                map(() => {
                    localStorage.removeItem('currentUser');
                    localStorage.removeItem('token');
                    this.currentUserSubject.next(null);
                    this.tokenSubject.next(null);
                })
            );
    }
}