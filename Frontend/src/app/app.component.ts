import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from './services/auth.service';
import { AuthFormComponent } from './components/auth-form/auth-form.component';
import { ChatComponent } from './components/chat/chat.component';

@Component({
  selector: 'app-root',
  template: `
    <ng-container *ngIf="!authService.isLoggedIn">
      <app-auth-form></app-auth-form>
    </ng-container>
    <ng-container *ngIf="authService.isLoggedIn">
      <app-chat></app-chat>
    </ng-container>
  `,
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    AuthFormComponent,
    ChatComponent
  ]
})
export class AppComponent {
  constructor(public authService: AuthService) { }
}