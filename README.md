# Door Control - Biometric Authentication POC

A proof-of-concept web application demonstrating biometric authentication using WebAuthn API with Go backend and server-side rendered HTML with HTMX.

## Features

- **Biometric Registration**: Register users with Face ID, Touch ID, or fingerprint scanner
- **Biometric Login**: Authenticate users with platform authenticators
- **WebAuthn API**: Secure public key cryptography
- **SQLite Database**: Stores user credentials and public keys
- **Server-Side Rendering**: Go templates with HTMX integration
- **Session Management**: Secure cookie-based sessions

## Technology Stack

- **Backend**: Go 1.21+
- **Database**: SQLite
- **Frontend**: HTML, CSS, Vanilla JavaScript
- **Authentication**: WebAuthn API
- **Libraries**:
  - `github.com/go-webauthn/webauthn` - WebAuthn implementation
  - `github.com/mattn/go-sqlite3` - SQLite driver
  - `github.com/gorilla/sessions` - Session management

## Project Structure

```
door-control/
├── main.go                 # Server entry point
├── go.mod                  # Go dependencies
├── go.sum                  # Go dependency checksums
├── door-control.db         # SQLite database (created on first run)
├── db/
│   ├── db.go              # Database operations
│   └── schema.sql         # Database schema
├── handlers/
│   ├── register.go        # Registration endpoints
│   ├── login.go           # Login endpoints
│   └── dashboard.go       # Protected dashboard
├── models/
│   └── user.go            # User model implementing WebAuthn interface
├── templates/
│   ├── register.html      # Registration page
│   ├── login.html         # Login page
│   └── dashboard.html     # Protected dashboard page
└── static/
    └── webauthn.js        # WebAuthn JavaScript helpers
```

## How It Works

### Registration Flow

1. User enters username and display name
2. User clicks "Register with Biometrics"
3. Backend generates a challenge and sends registration options to frontend
4. Frontend calls `navigator.credentials.create()` with WebAuthn options
5. Device prompts biometric authentication (Face ID/Touch ID/Fingerprint)
6. Device generates a key pair, stores private key in secure enclave
7. Public key and credential ID are returned to frontend
8. Frontend sends credential to backend
9. Backend verifies and stores: `user_id`, `credential_id`, `public_key`, `sign_count`
10. User is registered and session is created

### Authentication Flow

1. User enters username
2. User clicks "Login with Biometrics"
3. Backend generates authentication challenge and sends to frontend
4. Frontend calls `navigator.credentials.get()` with challenge
5. Device prompts biometric authentication
6. Device signs challenge with private key from secure enclave
7. Signed assertion is returned to frontend
8. Frontend sends assertion to backend
9. Backend verifies signature using stored public key
10. User is authenticated and session is created

## Installation & Setup

### Prerequisites

- Go 1.21 or higher
- A device with biometric authentication (or use Chrome DevTools Virtual Authenticator)

### Steps

1. **Build the application**:
   ```bash
   go build -o door-control
   ```

2. **Run the server**:
   ```bash
   ./door-control
   ```

3. **Access the application**:
   Open your browser and navigate to `http://localhost:8080`

## Testing Locally

### On macOS/iOS with Touch ID/Face ID

Simply access `http://localhost:8080` from Safari on your Mac or iPhone.

### On Android with Fingerprint

Access `http://localhost:8080` from Chrome on your Android device (may require port forwarding or ngrok for remote access).

### Using Chrome DevTools Virtual Authenticator

1. Open Chrome DevTools (F12)
2. Go to `⋮ More tools > WebAuthn`
3. Click "Enable virtual authenticator environment"
4. Click "Add authenticator"
5. Select "Internal Authenticator" with "User Verification: Yes"
6. Test registration and login flows

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Credentials Table
```sql
CREATE TABLE credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    credential_id BLOB UNIQUE NOT NULL,
    public_key BLOB NOT NULL,
    sign_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

## Security Considerations

⚠️ **This is a POC for demonstration purposes. For production use:**

1. **Change the session secret**: Update the secret key in `main.go`
2. **Enable HTTPS**: WebAuthn requires HTTPS in production (localhost exempt)
3. **Set Secure cookies**: Change `Secure: true` in session options for HTTPS
4. **Update RPID**: Set to your actual domain
5. **Add rate limiting**: Prevent brute force attacks
6. **Add CSRF protection**: Protect against CSRF attacks
7. **Add proper error handling**: Don't expose internal errors to users
8. **Implement proper session timeout**: Add automatic session expiration
9. **Add audit logging**: Track authentication attempts
10. **Use environment variables**: Store secrets in environment variables

## API Endpoints

- `GET /` - Redirects to login
- `GET /register` - Registration page
- `POST /register/begin` - Start registration flow
- `POST /register/finish` - Complete registration flow
- `GET /login` - Login page
- `POST /login/begin` - Start authentication flow
- `POST /login/finish` - Complete authentication flow
- `POST /logout` - Logout user
- `GET /dashboard` - Protected dashboard (requires authentication)

## Troubleshooting

### "User verification failed"
- Ensure your device has biometric authentication set up
- Check that biometric authentication is enabled in browser settings

### "Registration failed: Invalid RP ID"
- Make sure you're accessing via `localhost`, not `127.0.0.1`
- Check RPID matches your domain in `main.go`

### "Database locked" error
- Close other connections to the database
- Delete `door-control.db` and restart

## Browser Compatibility

- ✅ Chrome 67+
- ✅ Edge 18+
- ✅ Firefox 60+
- ✅ Safari 13+
- ✅ iOS Safari 14+
- ✅ Chrome Android 70+

## License

MIT

## Contributing

This is a proof-of-concept. Feel free to fork and improve!
