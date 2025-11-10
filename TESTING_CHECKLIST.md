# Authentication Testing Checklist

## Quick Test Steps

### 1. Check Authentication Flow
1. Open your browser and navigate to `http://localhost:4000`
2. You should see a "Login" button in the header
3. Click the "Login" button
4. You should be redirected to your OIDC provider (Keycloak)
5. Complete authentication
6. You should be redirected back to the home page
7. The header should now show your name/email and a "Logout" button

### 2. Verify Session Persistence
1. After logging in, refresh the page (F5)
2. You should remain logged in (no need to re-authenticate)
3. The header should still show your name/email

### 3. Test Project Creation
1. Once logged in, you should see a "New Project" button on the Projects page
2. Click the "New Project" button
3. A modal should appear
4. Enter a project name (e.g., "My Test Project")
5. Click "Create"
6. The project should be created and appear in the projects list

### 4. Test Logout
1. Click the "Logout" button in the header
2. You should be redirected to the home page
3. The header should now show the "Login" button again
4. The "New Project" button should not be visible

### 5. Browser Developer Tools Check

#### Check Session Cookie
1. Open Developer Tools (F12)
2. Go to Application → Cookies
3. You should see a cookie named `session_id`
4. Verify it has these properties:
   - HttpOnly: ✓
   - Secure: ✓ (if using HTTPS)
   - SameSite: Strict
   - Path: /

#### Check Network Requests
1. Open Developer Tools (F12)
2. Go to Network tab
3. After login, you should see:
   - `GET /auth/login` → 302 redirect to OIDC provider
   - `GET /auth/callback` → 302 redirect to home
   - `POST /auth/refresh` → 200 with user data
4. The `/auth/refresh` request should include the session cookie

#### Check Console Logs
1. Open Developer Tools (F12)
2. Go to Console tab
3. There should be no errors
4. You might see logs like:
   - "Failed to refresh session" (before login)
   - User data after successful refresh

## Troubleshooting

### Issue: "Login" button still shows after successful authentication

**Check:**
1. Open browser console and look for errors
2. Check Network tab for `/auth/refresh` request
3. Verify the response contains user data
4. Check Application → Cookies for `session_id` cookie

**Solution:**
- If `/auth/refresh` returns 401: Session expired or invalid
- If `/auth/refresh` returns 404: Route not configured correctly
- If no session cookie: Check COOKIE_SECURE setting (should be false for HTTP)

### Issue: Cannot create projects

**Check:**
1. Verify you are logged in (user name/email shows in header)
2. Check browser console for errors
3. Check Network tab for `/api/v1/projects` POST request

**Solution:**
- If 401 Unauthorized: Authentication middleware might need adjustment
- If 403 Forbidden: User might not have permissions

### Issue: Session does not persist after page refresh

**Check:**
1. Verify session cookie exists in Application → Cookies
2. Check cookie expiration time
3. Check COOKIE_SECURE setting

**Solution:**
- If using HTTP (localhost), set `COOKIE_SECURE=false` in .env
- If using HTTPS, set `COOKIE_SECURE=true` in .env

### Issue: Redirect loop during login

**Check:**
1. Verify `OIDC_REDIRECT_URL` matches registered redirect URI
2. Check backend logs for state validation errors
3. Verify OIDC provider configuration

**Solution:**
- Update `OIDC_REDIRECT_URL` to match the redirect URI registered with your OIDC provider
- Example: `OIDC_REDIRECT_URL=http://localhost:4000/auth/callback`

## Backend Logs

To view backend logs for debugging:

```bash
docker logs -f librecov-librecov-1
```

Look for:
- "Session created for user X: <session_id>" - Successful login
- "State mismatch" - CSRF validation failed
- "Failed to exchange token" - OIDC token exchange failed
- "User found: <email>" - User retrieved from database

## Current Status Based on Logs

From your recent logs, I can see:
- ✅ Login flow is working
- ✅ Session creation is successful
- ✅ `/auth/refresh` is now working correctly (returning 200)
- ✅ User data is being retrieved

**Expected Behavior:**
After the latest rebuild, the authentication should now work correctly. The header should show your user information after login, and you should be able to create new projects.

## Next Steps

1. **Clear your browser cache and cookies** for localhost:4000
2. **Refresh the page** (Ctrl+F5 or Cmd+Shift+R)
3. **Click Login** and complete authentication
4. **Verify** that the header shows your name/email
5. **Try creating a project** using the "New Project" button

If you still see issues, check the browser console and network tab for specific error messages.
