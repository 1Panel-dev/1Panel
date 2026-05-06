# API Usage Examples - Multi-User System

## Authentication

### Login with credentials
```bash
curl -X POST http://localhost:8080/api/v2/core/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "name": "admin",
    "password": "your_password",
    "language": "en"
  }'
```

Response:
```json
{
  "code": 200,
  "data": {
    "name": "admin",
    "token": "session_token",
    "mfaStatus": "disable",
    "role": "admin"
  }
}
```

## User Management

### Create a new user
```bash
curl -X POST http://localhost:8080/api/v2/core/users \
  -H "Content-Type: application/json" \
  -H "Cookie: SESSIONID=your_token" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "secure_password_123",
    "role": "reseller",
    "realName": "John Doe",
    "phone": "+1234567890",
    "remark": "Reseller account"
  }'
```

### List all users
```bash
curl -X GET "http://localhost:8080/api/v2/core/users?pageNum=1&pageSize=10&role=reseller" \
  -H "Cookie: SESSIONID=your_token"
```

Response:
```json
{
  "code": 200,
  "data": {
    "total": 5,
    "items": [
      {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "role": "admin",
        "status": "active",
        "realName": "Administrator",
        "phone": "",
        "lastLogin": 1715000000,
        "remark": "",
        "createdAt": 1714900000,
        "updatedAt": 1714950000
      },
      {
        "id": 2,
        "username": "john_doe",
        "email": "john@example.com",
        "role": "reseller",
        "status": "active",
        "realName": "John Doe",
        "phone": "+1234567890",
        "lastLogin": 1714990000,
        "remark": "Reseller account",
        "createdAt": 1714900100,
        "updatedAt": 1714950100
      }
    ]
  }
}
```

### Get user details
```bash
curl -X GET http://localhost:8080/api/v2/core/users/2 \
  -H "Cookie: SESSIONID=your_token"
```

Response:
```json
{
  "code": 200,
  "data": {
    "user": {
      "id": 2,
      "username": "john_doe",
      "email": "john@example.com",
      "role": "reseller",
      "status": "active",
      "realName": "John Doe",
      "phone": "+1234567890",
      "lastLogin": 1714990000,
      "remark": "Reseller account",
      "createdAt": 1714900100,
      "updatedAt": 1714950100
    },
    "permissions": [
      "user:view",
      "user:create",
      "user:update",
      "user:delete",
      "host:view",
      "host:monitor",
      "app:manage",
      "app:create",
      "database:manage"
    ]
  }
}
```

### Update user
```bash
curl -X PUT http://localhost:8080/api/v2/core/users \
  -H "Content-Type: application/json" \
  -H "Cookie: SESSIONID=your_token" \
  -d '{
    "id": 2,
    "email": "newemail@example.com",
    "role": "user",
    "status": "active",
    "realName": "John Doe Updated",
    "phone": "+9876543210",
    "remark": "Updated reseller account"
  }'
```

### Delete user
```bash
curl -X DELETE http://localhost:8080/api/v2/core/users/2 \
  -H "Cookie: SESSIONID=your_token"
```

## Password Management

### Change own password
```bash
curl -X POST http://localhost:8080/api/v2/core/users/password/change \
  -H "Content-Type: application/json" \
  -H "Cookie: SESSIONID=your_token" \
  -d '{
    "userId": 2,
    "oldPassword": "old_password_123",
    "newPassword": "new_password_456"
  }'
```

### Reset user password (admin only)
```bash
curl -X POST http://localhost:8080/api/v2/core/users/password/reset \
  -H "Content-Type: application/json" \
  -H "Cookie: SESSIONID=admin_token" \
  -d '{
    "userId": 2,
    "newPassword": "temporary_password_789"
  }'
```

## Permissions Management

### Get user permissions
```bash
curl -X GET http://localhost:8080/api/v2/core/users/2/permissions \
  -H "Cookie: SESSIONID=your_token"
```

Response:
```json
{
  "code": 200,
  "data": [
    "user:view",
    "user:create",
    "user:update",
    "user:delete",
    "host:view",
    "host:monitor",
    "host:manage",
    "app:manage",
    "app:create",
    "app:update",
    "app:delete",
    "app:view",
    "app:install",
    "app:uninstall"
  ]
}
```

### Assign permissions to user
```bash
curl -X POST http://localhost:8080/api/v2/core/users/permissions \
  -H "Content-Type: application/json" \
  -H "Cookie: SESSIONID=admin_token" \
  -d '{
    "userId": 2,
    "permissions": [
      "user:view",
      "app:view",
      "app:install",
      "database:view",
      "host:monitor"
    ]
  }'
```

## Profile Management

### Get current user profile
```bash
curl -X GET http://localhost:8080/api/v2/core/users/profile \
  -H "Cookie: SESSIONID=your_token"
```

Response:
```json
{
  "code": 200,
  "data": {
    "user": {
      "id": 2,
      "username": "john_doe",
      "email": "john@example.com",
      "role": "reseller",
      "status": "active",
      "realName": "John Doe",
      "phone": "+1234567890",
      "lastLogin": 1714990000,
      "remark": "Reseller account",
      "createdAt": 1714900100,
      "updatedAt": 1714950100
    },
    "permissions": [
      "user:view",
      "app:manage",
      "database:view"
    ]
  }
}
```

## Login History

### Get login history for user
```bash
curl -X GET http://localhost:8080/api/v2/core/users/2/login-history \
  -H "Cookie: SESSIONID=your_token"
```

Response:
```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "userId": 2,
      "ip": "192.168.1.100",
      "address": "New York, USA",
      "agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
      "status": "success",
      "message": "Login successful",
      "loginAt": 1714990000,
      "createdAt": 1714990000,
      "updatedAt": 1714990000
    },
    {
      "id": 2,
      "userId": 2,
      "ip": "192.168.1.101",
      "address": "New York, USA",
      "agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0)",
      "status": "success",
      "message": "Login successful",
      "loginAt": 1714985000,
      "createdAt": 1714985000,
      "updatedAt": 1714985000
    }
  ]
}
```

## Available Permissions

### User Management
- `user:view` - View user information
- `user:create` - Create new users
- `user:update` - Update user information
- `user:delete` - Delete users
- `user:manage` - Full user management
- `user:password` - Change password

### Host/Node Management
- `host:view` - View hosts
- `host:monitor` - Monitor host resources
- `host:manage` - Full host management
- `host:create` - Create host connections
- `host:update` - Update host information
- `host:delete` - Delete hosts

### Application Management
- `app:view` - View applications
- `app:create` - Create applications
- `app:update` - Update applications
- `app:delete` - Delete applications
- `app:manage` - Full app management
- `app:install` - Install applications
- `app:uninstall` - Uninstall applications

### Database Management
- `database:view` - View databases
- `database:create` - Create databases
- `database:update` - Update databases
- `database:delete` - Delete databases
- `database:manage` - Full database management
- `database:backup` - Create database backups

### Website Management
- `website:view` - View websites
- `website:create` - Create websites
- `website:update` - Update websites
- `website:delete` - Delete websites
- `website:manage` - Full website management

### Backup Management
- `backup:view` - View backups
- `backup:create` - Create backups
- `backup:delete` - Delete backups
- `backup:manage` - Full backup management

### System Settings
- `setting:view` - View settings
- `setting:manage` - Manage settings
- `system:manage` - Full system management
- `system:upgrade` - System upgrades
- `system:log` - View system logs
- `system:restart` - Restart system

---

## HTTP Status Codes

- `200` - Success
- `400` - Bad request / Validation error
- `401` - Unauthorized / Not authenticated
- `403` - Forbidden / Insufficient permissions
- `404` - Not found
- `500` - Internal server error

## Error Responses

```json
{
  "code": 401,
  "message": "ErrNotLogin"
}
```

```json
{
  "code": 403,
  "message": "insufficient permissions"
}
```

```json
{
  "code": 400,
  "message": "ErrUserAlreadyExists"
}
```

---

## Role-Based API Access

### Admin Can Access:
- All user management endpoints
- All permission endpoints
- Password reset for any user
- All system endpoints

### Reseller Can Access:
- User viewing and management
- App management endpoints
- Database endpoints
- Website endpoints
- Backup endpoints
- Own profile and login history

### User Can Access:
- Own profile
- Password change (own only)
- View-only endpoints
- Login history (own only)

---

## Notes

1. All requests require a valid session cookie (SESSIONID)
2. Role and permission validation is performed on the backend
3. Failed attempts are logged and tracked by IP
4. Admin role ("admin") has access to all features by default
5. Custom permissions can be assigned per user
6. Password changes require old password verification
7. Admin can reset passwords without old password
