import { Component } from "@angular/core";
import { AuthService } from "../../services/auth.service";
import { CommonModule } from "@angular/common";
import { FormsModule } from "@angular/forms";

@Component({
    selector: 'app-auth-form',
    templateUrl: 'auth-form.component.html',
    styleUrl: 'auth-form.component.scss',
    standalone: true,
    imports: [CommonModule, FormsModule]
})
export class AuthFormComponent {
    username: string = '';
    password: string = '';
    error: string = '';

    constructor(private authService: AuthService) { }

    onSubmit(isLogin: boolean): void {
        const credentials = { username: this.username, password: this.password };

        this.error = '';

        if (isLogin) {
            this.authService.login(credentials).subscribe({
                error: (err) => this.error = err.error || 'Login failed'
            });
        } else {
            this.authService.register(credentials).subscribe({
                error: (err) => this.error = err.error || 'Registration failed'
            });
        }
    }
}