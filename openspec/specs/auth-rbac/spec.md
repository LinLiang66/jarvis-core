## ADDED Requirements

### Requirement: Userinfo returns roles and permissions

`GET /api/auth/userinfo` MUST return user profile plus `roles: string[]` (role codes) and `permissions: string[]`. Users with `role_admin` MUST receive `permissions` containing `*:*:*`. For non-super-admin users, `permissions` MUST be the distinct union of `permission` values from enabled (`status='1'`) menu nodes of type 3 (and type 2 when `permission` is set) that are authorized to the user via `sys_role_menu` for any of the user's enabled roles.

#### Scenario: Super admin userinfo

- **WHEN** a user with `role_admin` requests userinfo
- **THEN** `roles` includes `role_admin` and `permissions` includes `*:*:*`

#### Scenario: Multi-role user permissions from menus

- **WHEN** a user has roles `role_user` and `operator` (both enabled) and `sys_role_menu` grants button permissions `crud:list` and `system:user:add`
- **THEN** `permissions` includes both `crud:list` and `system:user:add`

### Requirement: Login response alignment

`POST /api/auth/login` success payload MUST include `user` object compatible with userinfo (including `roles`; `permissions` MAY be omitted on login if userinfo is fetched immediately after).

#### Scenario: Login then permission store

- **WHEN** frontend completes login and calls `fetchUserInfo`
- **THEN** `usePermissionStore.setRoles` and `setPermissions` are invoked with API values

### Requirement: Menu routes filtered by roles

`GET /api/menu/routes` MUST authorize menu visibility using `sys_role_menu`: a type-1 or type-2 menu is visible if it is directly linked to any of the user's enabled roles, or is an ancestor of such a linked menu. Super admin (`role_admin`) MUST receive all enabled route menus without consulting `sys_role_menu`. The menu item field `roles` MUST NOT be used for filtering.

#### Scenario: User sees menus granted by role_menu

- **WHEN** menu id `11` is linked to role `role_user` and the user has `role_user`
- **THEN** menu `11` and its ancestors appear in `/api/menu/routes`

#### Scenario: User does not see unlinked menus

- **WHEN** menu id `12` is not linked to any of the user's enabled roles
- **THEN** menu `12` is not included in `/api/menu/routes`

#### Scenario: Super admin sees all

- **WHEN** user has `role_admin`
- **THEN** all enabled route menus (type 1 and 2) are returned

### Requirement: Super admin guard on write APIs

User and role mutation endpoints MUST require super-admin (holder of `role_admin` or equivalent guard replacing `role == "admin"`).

#### Scenario: Normal user cannot create user

- **WHEN** a user without `role_admin` calls `POST /api/user`
- **THEN** the system returns HTTP 403

### Requirement: System management menu structure

The menu data source (database seed, not MOCK) MUST include parent **系统管理** (`/system`) with children:

- 用户管理 → `component: system/user/index`, path `/system/user/index`
- 角色管理 → `component: system/role/index`, path `/system/role/index`
- 菜单管理 → `component: system/menu/index`, path `/system/menu/index`

#### Scenario: Routes registered after login

- **WHEN** super admin logs in and `generateRoutes` runs
- **THEN** sidebar shows 系统管理 with 用户管理, 角色管理, and 菜单管理 entries

### Requirement: Frontend permission utilities wired

After login or `fetchUserInfo`, frontend MUST populate `usePermissionStore` so `hasRole` and `hasPerm` work with `SUPER_ADMIN_ROLE` and `SUPER_ADMIN_PERMISSION` from `@/core/config`.

#### Scenario: Button hidden without permission

- **WHEN** a button uses `v-hasPerm="'system:user:add'"` and user lacks that permission
- **THEN** the button is not rendered

### Requirement: Frontend refresh routes after permission changes

The frontend MUST expose `refreshRoutes()` that: removes previously registered dynamic routes, calls `fetchUserInfo` to refresh `permissions`, fetches `/api/menu/routes`, and re-registers routes via `setRoutes`. After menu CRUD success or role menu assignment success, the UI MUST invoke `refreshRoutes()` automatically for the current session.

On login and logout, the frontend MUST reset the route-loaded flag and clear persisted dynamic routes so a different user receives the correct menu without re-login artifacts.

#### Scenario: Menu save triggers refresh

- **WHEN** super admin saves a menu in 菜单管理 and the API returns success
- **THEN** the frontend calls `refreshRoutes` without requiring re-login

#### Scenario: Manual refresh available

- **WHEN** super admin clicks 刷新路由 on the menu management page
- **THEN** `refreshRoutes` runs and the sidebar reflects the latest menu tree

#### Scenario: Role menu save triggers refresh

- **WHEN** super admin saves role menu permissions and the API returns success
- **THEN** the frontend calls `refreshRoutes` and updates `usePermissionStore.permissions`

#### Scenario: Switch user reloads menu

- **WHEN** user A logs out and user B logs in within the same browser session
- **THEN** the sidebar shows user B's authorized menus only

### Requirement: Deprecated single role field

The API MUST NOT rely on `user.role` as a single string for authorization after migration. **BREAKING** for clients expecting only `role: string`.

#### Scenario: Old client reads userinfo

- **WHEN** client expects `role: "admin"`
- **THEN** client MUST migrate to `roles: string[]` (documented in change proposal)
