# Termix Integration

SSHBuddy can automatically import hosts from your Termix server.

## Configuration

1. Copy the example config:
   ```bash
   cp termix.example.json ~/.config/sshbuddy/termix.json
   ```

2. Edit `~/.config/sshbuddy/termix.json`:
   ```json
   {
     "enabled": true,
     "baseUrl": "https://your-termix-server.com/api",
     "username": "your-username",
     "password": "your-password"
   }
   ```

   **Important:** The `baseUrl` should be the base API URL (e.g., `https://example.com/api`), not including the specific endpoints like `/users/login` or `/ssh/db/host`.

## How It Works

When enabled, SSHBuddy will:
1. Authenticate with your Termix server using the credentials
2. Fetch all hosts from `/ssh/db/host` endpoint
3. Display them in the TUI with `source: termix` label
4. Cache the JWT token for future requests

## Troubleshooting

If you encounter errors:

1. **Check the debug logs:**
   ```bash
   ./view-logs.sh
   # or
   cat /tmp/sshbuddy-debug.log
   ```

2. **Test your credentials manually:**
   ```bash
   # Test authentication
   curl --request POST \
     --url https://YOUR_BASE_URL/users/login \
     --header 'Content-Type: application/json' \
     --data '{"username": "YOUR_USERNAME","password": "YOUR_PASSWORD"}'
   
   # Should return 200 with Set-Cookie: jwt=...
   ```

3. **Test host fetching:**
   ```bash
   # Replace JWT_TOKEN with the token from above
   curl --request GET \
     --url https://YOUR_BASE_URL/ssh/db/host \
     --header 'Cookie: jwt=JWT_TOKEN; i18nextLng=en'
   ```

4. **Common issues:**
   - **Invalid JSON error**: Check that `baseUrl` is correct (should end with `/api`, not include `/users/login`)
   - **Authentication failed**: Verify username and password
   - **Connection failed**: Check network connectivity and that the server is accessible

## Disabling Termix

To disable Termix integration, set `"enabled": false` in your config:
```json
{
  "enabled": false,
  "baseUrl": "...",
  "username": "...",
  "password": "..."
}
```

Or delete the config file:
```bash
rm ~/.config/sshbuddy/termix.json
```
