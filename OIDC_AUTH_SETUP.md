# OIDC Authentication - Environment Variables

This document describes the required environment variables for the secure OIDC authentication implementation.

## Required Environment Variables

### Backend Configuration

```bash
# OIDC Provider Configuration
OIDC_ISSUER=https://your-oidc-provider.com
OIDC_CLIENT_ID=your-client-id
OIDC_REDIRECT_URL=https://your-app.com/auth/callback

# Cookie Security Settings
COOKIE_DOMAIN=your-app.com  # Leave empty for current domain
COOKIE_SECURE=true          # Set to true in production with HTTPS

# Frontend URL (for redirect after successful login)
FRONTEND_URL=https://your-app.com

# Server Port (optional, defaults to 4000)
PORT=4000

# Database Configuration (existing)
DB_HOST=postgres
DB_PORT=5432
DB_USER=librecov
DB_PASSWORD=your-password
DB_NAME=librecov
```

## Security Features Implemented

### 1. **PKCE (Proof Key for Code Exchange)**
   - Generates a cryptographically secure verifier and challenge
   - Protects against authorization code interception attacks
   - Essential for public clients (no client secret)

### 2. **State Parameter (CSRF Protection)**
   - Cryptographically secure random state value
   - Stored server-side in session store
   - Prevents CSRF attacks during OIDC flow

### 3. **Secure Session Management**
   - Session cookies with HttpOnly flag (prevents XSS)
   - Secure flag for HTTPS (prevents MITM)
   - SameSite=Strict for session cookies (prevents CSRF)
   - SameSite=Lax for state cookies (allows OIDC redirects)
   - Server-side session storage (not in cookies)

### 4. **Session Expiration & Cleanup**
   - Sessions expire after 24 hours
   - State data expires after 10 minutes
   - Automatic cleanup routine runs every 5 minutes
   - Frontend refreshes session every 15 minutes

## Authentication Flow

### 1. Login Initiation (`/auth/login`)
   - Generate state token (CSRF protection)
   - Generate PKCE verifier and challenge
   - Store state and verifier server-side
   - Set state cookie with SameSite=Lax
   - Redirect to OIDC provider with challenge

### 2. OIDC Callback (`/auth/callback`)
   - Verify state parameter matches cookie
   - Retrieve PKCE verifier from session store
   - Exchange authorization code with PKCE verifier
   - Verify ID token
   - Extract user claims
   - Create or update user in database
   - Create session and set secure cookie
   - Redirect to frontend

### 3. Session Refresh (`/auth/refresh`)
   - Frontend calls periodically (every 15 minutes)
   - Validates session cookie
   - Returns current user information
   - Extends session lifetime

### 4. Logout (`/auth/logout`)
   - Delete session from store
   - Clear session cookie
   - Redirect to home page

## API Endpoints

- `GET /auth/config` - Get authentication configuration (public)
- `GET /auth/login` - Initiate OIDC login flow
- `GET /auth/callback` - Handle OIDC callback
- `POST /auth/refresh` - Refresh session and get user info
- `POST /auth/logout` - Logout and clear session
- `GET /auth/me` - Get current user (requires auth)

## Frontend Integration

The frontend automatically:
- Checks for existing session on app initialization
- Refreshes session every 15 minutes
- Handles session expiration gracefully
- Redirects to login when needed

No token management in localStorage - all authentication is cookie-based.

## Production Deployment Notes

1. **HTTPS Required**: Set `COOKIE_SECURE=true` in production
2. **Cookie Domain**: Set `COOKIE_DOMAIN` to your domain or leave empty
3. **OIDC Provider**: Register `OIDC_REDIRECT_URL` with your provider
4. **Session Storage**: Consider using Redis or database for production (currently in-memory)
5. **CORS**: Configure CORS properly for your domain

## Testing

Test the authentication flow:

1. Navigate to `/auth/login`
2. Complete OIDC authentication
3. Should redirect back to home page with session established
4. Check browser dev tools for session cookie
5. Verify session refresh works in network tab

## Troubleshooting

Check backend logs for detailed error messages:
- State validation issues
- Token exchange failures
- ID token verification errors
- Session creation/retrieval issues

All critical operations are logged with context for debugging.
