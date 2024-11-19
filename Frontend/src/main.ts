import { bootstrapApplication } from '@angular/platform-browser';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { AppComponent } from './app/app.component';
import { AuthInterceptor } from './app/interceptors/auth.interceptor';
import { inject } from '@angular/core';
import { AuthService } from './app/services/auth.service';
import { WebsocketService } from './app/services/websocket.service';

bootstrapApplication(AppComponent, {
  providers: [
    provideHttpClient(
      withInterceptors([
        (req, next) => {
          const authInterceptor = new AuthInterceptor(
            inject(AuthService),
            inject(WebsocketService)
          );
          return authInterceptor.intercept(req, { handle: next });
        }
      ])
    )
  ]
}).catch(err => console.error(err));