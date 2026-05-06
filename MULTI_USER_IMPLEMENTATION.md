# Multi-User System with Role-Based Access Control Implementation

## Overview
This document describes the multi-user system that has been implemented for 1Panel, featuring a three-tier user role system with comprehensive permission management.

## User Roles

### 1. **Main Admin (admin)**
- Full access to all system features
- Can manage all users and permissions
- Can modify system settings
- Can perform system upgrades and restarts
- Complete access to all resources

#### Permissions:
- `admin:all` - Grants all permissions

### 2. **Reseller (reseller)**
- Can manage applications and services for their accounts
- Can create and manage sub-users
- Can monitor host/server resources
- Limited system settings access
- Can manage databases, websites, and backups

#### Permissions:
- `user:view`, `user:create`, `user:update`, `user:delete`
- `host:view`, `host:monitor`, `host:manage`
- `app:*` - Full app management
- `database:*` - Full database management
- `website:*` - Full website management
- `backup:*` - Full backup management
- `setting:view`

### 3. **User (user)**
- Basic access to view their own resources
- Can only change their own password
- Limited viewing permissions for apps and databases
- Cannot create or manage other users

#### Permissions:
- `user:view`, `user:password` - Own profile only
- `host:monitor`, `host:view` - Monitor only
- `app:view`, `app:install` - View and install only
- `database:view` - View only
- `website:view` - View only
- `backup:view` - View only
- `setting:view` - View only

## Database Schema

### User Table (`users`)
```sql
CREATE TABLE users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(255) NOT NULL UNIQUE,
  email VARCHAR(255),
  password VARCHAR(255) NOT NULL,
  role VARCHAR(50) NOT NULL DEFAULT 'user',
  status VARCHAR(50) NOT NULL DEFAULT 'active',
  parent_id INT,
  real_name VARCHAR(255),
  phone VARCHAR(20),
  last_login TIMESTAMP,
  remark TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)
```

### User Permissions Table (`user_permissions`)
```sql
CREATE TABLE user_permissions (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  permission VARCHAR(255) NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
)
```

### User Login History Table (`user_login_histories`)
```sql
CREATE TABLE user_login_histories (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  ip VARCHAR(50),
  address VARCHAR(255),
  agent TEXT,
  status VARCHAR(50) NOT NULL,
  message TEXT,
  login_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
)
```

## API Endpoints

### User Management
- `GET /api/v2/core/users` - List all users
- `POST /api/v2/core/users` - Create new user
- `GET /api/v2/core/users/:id` - Get user details
- `PUT /api/v2/core/users` - Update user
- `DELETE /api/v2/core/users/:id` - Delete user

### Password Management
- `POST /api/v2/core/users/password/change` - Change own password
- `POST /api/v2/core/users/password/reset` - Reset user password (admin only)

### Permissions
- `GET /api/v2/core/users/:id/permissions` - Get user permissions
- `POST /api/v2/core/users/permissions` - Assign permissions

### Profile
- `GET /api/v2/core/users/profile` - Get current user profile
- `GET /api/v2/core/users/:id/login-history` - Get login history

## Authentication Flow

1. User submits login credentials (username/password)
2. System checks User table first (new multi-user system)
3. Falls back to settings-based auth for backward compatibility
4. Validates security entrance if configured
5. Triggers MFA if enabled
6. Creates session with user role and permissions
7. Sets session cookies with user context

## Middleware

### SessionAuth Middleware
- Located in `core/middleware/session.go`
- Validates active session
- Refreshes session if needed
- Checks session timeout

### UserAuthMiddleware
- Located in `core/middleware/role.go`
- Adds user context to request
- Retrieves user permissions
- Sets UserID, UserRole, and Permissions in context

### RoleAuth Middleware
- Located in `core/middleware/role.go`
- Enforces role-based access control
- Checks if user has required role
- Returns 401 if user lacks required role

### PermissionAuth Middleware
- Located in `core/middleware/role.go`
- Enforces permission-based access control
- Checks specific permissions
- Allows admin bypass

## Service Layer

### UserService (`core/app/service/user.go`)
Key methods:
- `CreateUser(req)` - Create new user with role
- `UpdateUser(req)` - Update user information
- `GetUserByID(id)` - Retrieve user details
- `GetUserByUsername(username)` - Find user by username
- `GetUserList(req)` - List users with pagination
- `DeleteUser(id)` - Delete user (prevents main admin deletion)
- `ChangePassword(req)` - User changes own password
- `ResetPassword(req)` - Admin resets user password
- `AssignPermissions(req)` - Assign custom permissions
- `GetUserPermissions(userID)` - Get user permissions
- `HasPermission(userID, permission)` - Check single permission
- `CheckUserRole(userID, roles...)` - Check if user has role

## Repository Layer

### UserRepository (`core/app/repo/user_repo.go`)
Handles all database operations:
- User CRUD operations
- Permission management
- Login history tracking

## Data Transfer Objects (DTOs)

### CreateUserRequest
```go
type CreateUserRequest struct {
  Username string
  Email    string
  Password string
  Role     string // user, reseller, admin
  RealName string
  Phone    string
  Remark   string
}
```

### UserResponse
```go
type UserResponse struct {
  ID        uint
  Username  string
  Email     string
  Role      string
  Status    string
  RealName  string
  Phone     string
  LastLogin int64
  Remark    string
  CreatedAt int64
  UpdatedAt int64
}
```

### UserDetailResponse
```go
type UserDetailResponse struct {
  User        UserResponse
  Permissions []string
}
```

## Migration

The database migration (`AddUserTables` in `core/init/migration/migrations/init.go`):
1. Creates user-related tables
2. Checks for existing admin user
3. Migrates existing single-user setup to admin user if needed
4. Maintains backward compatibility

## Backward Compatibility

The implementation maintains full backward compatibility:
1. Old settings-based authentication still works
2. Existing credentials are migrated to admin user on first run
3. Session-based authentication continues to work
4. All existing APIs remain functional

## Security Features

1. **Password Encryption** - Passwords are encrypted using system encryption utilities
2. **Session Management** - Session-based authentication with timeout
3. **MFA Support** - Multi-factor authentication available
4. **Login History** - Tracks all login attempts
5. **Permission Validation** - Every action is permission-checked
6. **IP Tracking** - Failed login attempts tracked by IP

## Future Enhancements

1. **Resource Scoping** - Scope users to specific hosts/resources
2. **API Keys** - Support API-based authentication
3. **LDAP/OAuth** - Integration with external auth systems
4. **Audit Logging** - Detailed audit logs for compliance
5. **Advanced Permissions** - More granular permission system
6. **2FA** - Two-factor authentication
7. **SSO** - Single sign-on integration

## Files Created/Modified

### New Files:
- `core/constant/role.go` - Role and permission constants
- `core/app/model/user.go` - User models
- `core/app/dto/user.go` - User DTOs
- `core/app/repo/user_repo.go` - User repository
- `core/app/service/user.go` - User service
- `core/app/api/v2/user.go` - User API endpoints
- `core/middleware/role.go` - Role middleware
- `core/router/ro_user.go` - User router
- `core/init/migration/migrations/user_tables.go` (content in init.go)

### Modified Files:
- `core/init/migration/migrate.go` - Added user tables migration
- `core/init/migration/migrations/init.go` - Added AddUserTables migration
- `core/app/service/auth.go` - Updated Login and MFALogin methods
- `core/app/dto/auth.go` - Updated UserLoginInfo with role field
- `core/app/api/v2/helper/helper.go` - Added user helper functions
- `core/app/api/v2/entry.go` - Registered UserApi
- `core/router/common.go` - Added UserRouter

## Testing Recommendations

1. **User Creation** - Test creating users with different roles
2. **Permission Checking** - Verify permission enforcement
3. **Role Switching** - Test different role operations
4. **Backward Compatibility** - Ensure old auth still works
5. **Middleware** - Test role and permission middleware
6. **Login History** - Verify login tracking
7. **Password Management** - Test password change/reset

## Deployment Notes

1. Run database migrations automatically on startup
2. Existing single-user systems will have admin user created
3. No manual intervention needed for upgrade
4. All existing sessions remain valid
5. New users can be created through API

---

## Implementation by Features

### Multi-User Support ✅
- Multiple users can access the system
- Each user has independent credentials
- User management interface available

### Role-Based Access Control ✅
- Three tiers: User, Reseller, Admin
- Role-based capabilities
- Permission system for fine-grained control

### User Management ✅
- Create/update/delete users
- Password reset capability
- User listing and search

### Permission Management ✅
- Role-based default permissions
- Custom permission assignment
- Permission validation on endpoints

### Login Tracking ✅
- Login history per user
- Last login timestamp
- Failed login tracking by IP

---

**Implementation Date:** May 6, 2026
**Status:** Complete
