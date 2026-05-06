package constant

// User Roles
const (
	RoleUser      = "user"
	RoleReseller  = "reseller"
	RoleAdminMain = "admin"
)

// User Status
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)

// Permission Actions
const (
	// Admin permissions
	PermissionAdminAll = "admin:all"

	// User management permissions
	PermissionUserManage   = "user:manage"
	PermissionUserCreate   = "user:create"
	PermissionUserUpdate   = "user:update"
	PermissionUserDelete   = "user:delete"
	PermissionUserView     = "user:view"
	PermissionUserPassword = "user:password"

	// Host/Node management
	PermissionHostManage   = "host:manage"
	PermissionHostCreate   = "host:create"
	PermissionHostUpdate   = "host:update"
	PermissionHostDelete   = "host:delete"
	PermissionHostView     = "host:view"
	PermissionHostMonitor  = "host:monitor"

	// App management
	PermissionAppManage   = "app:manage"
	PermissionAppCreate   = "app:create"
	PermissionAppUpdate   = "app:update"
	PermissionAppDelete   = "app:delete"
	PermissionAppView     = "app:view"
	PermissionAppInstall  = "app:install"
	PermissionAppUninstall = "app:uninstall"

	// Database management
	PermissionDatabaseManage   = "database:manage"
	PermissionDatabaseCreate   = "database:create"
	PermissionDatabaseUpdate   = "database:update"
	PermissionDatabaseDelete   = "database:delete"
	PermissionDatabaseView     = "database:view"
	PermissionDatabaseBackup   = "database:backup"

	// Website management
	PermissionWebsiteManage   = "website:manage"
	PermissionWebsiteCreate   = "website:create"
	PermissionWebsiteUpdate   = "website:update"
	PermissionWebsiteDelete   = "website:delete"
	PermissionWebsiteView     = "website:view"

	// Backup management
	PermissionBackupManage = "backup:manage"
	PermissionBackupCreate = "backup:create"
	PermissionBackupDelete = "backup:delete"
	PermissionBackupView   = "backup:view"

	// Settings management
	PermissionSettingManage = "setting:manage"
	PermissionSettingView   = "setting:view"

	// System settings
	PermissionSystemManage    = "system:manage"
	PermissionSystemUpgrade   = "system:upgrade"
	PermissionSystemLog       = "system:log"
	PermissionSystemRestart   = "system:restart"
)

// Role Permissions mapping
var RolePermissions = map[string][]string{
	RoleAdminMain: {
		PermissionAdminAll,
	},
	RoleReseller: {
		// User management (limited to their sub-users)
		PermissionUserView,
		PermissionUserCreate,
		PermissionUserUpdate,
		PermissionUserDelete,

		// Host/Node viewing and monitoring
		PermissionHostView,
		PermissionHostMonitor,
		PermissionHostManage,

		// App management
		PermissionAppManage,
		PermissionAppCreate,
		PermissionAppUpdate,
		PermissionAppDelete,
		PermissionAppView,
		PermissionAppInstall,
		PermissionAppUninstall,

		// Database management
		PermissionDatabaseManage,
		PermissionDatabaseCreate,
		PermissionDatabaseUpdate,
		PermissionDatabaseDelete,
		PermissionDatabaseView,
		PermissionDatabaseBackup,

		// Website management
		PermissionWebsiteManage,
		PermissionWebsiteCreate,
		PermissionWebsiteUpdate,
		PermissionWebsiteDelete,
		PermissionWebsiteView,

		// Backup management
		PermissionBackupManage,
		PermissionBackupCreate,
		PermissionBackupDelete,
		PermissionBackupView,

		// Settings (limited)
		PermissionSettingView,
	},
	RoleUser: {
		// User can only view their own profile
		PermissionUserView,
		PermissionUserPassword,

		// Basic host monitoring
		PermissionHostMonitor,
		PermissionHostView,

		// App viewing and install (limited)
		PermissionAppView,
		PermissionAppInstall,

		// Database viewing
		PermissionDatabaseView,

		// Website viewing
		PermissionWebsiteView,

		// Backup viewing
		PermissionBackupView,

		// Settings viewing
		PermissionSettingView,
	},
}
