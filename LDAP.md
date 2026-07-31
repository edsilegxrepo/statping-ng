# LDAP Authentication

Statping-ng supports LDAP/Active Directory authentication for enterprise environments.

## Features

- **LDAPS (Implicit TLS)**: Automatic on ports 636 (domain) and 3269 (global catalog)
- **StartTLS**: Optional upgrade for ports 389 and 3268
- **Template Presets**: OpenLDAP, Microsoft Active Directory, FreeIPA
- **Authorized Group**: Restrict application access to specific group members
- **User Auto-Provisioning**: LDAP users are created automatically on first login
- **Test Connection**: Verify LDAP connectivity before saving settings

## Connection Settings

| Setting | Description |
|---------|-------------|
| LDAP Host | LDAP server hostname (e.g., `ldap.example.com`) |
| LDAP Port | `636` or `3269` for LDAPS, `389` or `3268` for StartTLS |
| Use StartTLS | Upgrade plain connection to TLS (ports 389/3268 only) |
| Skip TLS Verify | Skip certificate verification (not recommended for production) |

### Port Behavior

- **Port 636/3269**: Implicit TLS is automatically enabled (LDAPS)
- **Port 389/3268**: Plain connection by default, use StartTLS to upgrade to TLS

## Service Account (Bind DN)

A service account is required to search for users in the directory.

| Setting | Description |
|---------|-------------|
| Bind DN | Service account distinguished name (e.g., `cn=service,dc=example,dc=com`) |
| Bind Password | Service account password |

## User Search Settings

| Setting | Description | OpenLDAP | Active Directory | FreeIPA |
|---------|-------------|----------|------------------|---------|
| Base DN | Search base (e.g., `dc=example,dc=com`) | - | - | - |
| User Filter | LDAP filter with `%s` placeholder | `(&(objectClass=inetOrgPerson)(uid=%s))` | `(&(objectClass=user)(sAMAccountName=%s))` | `(&(objectClass=person)(uid=%s))` |
| Username Attribute | Username field | `uid` | `sAMAccountName` | `uid` |
| Email Attribute | Email field | `mail` | `mail` | `mail` |

## Authorization

### Authorized Group

When enabled, users must be a member of the specified group to access the application.

- **Disabled (default)**: Any authenticated LDAP user can log in
- **Enabled**: Only members of the specified group(s) can log in
- **Multiple Groups**: Comma-separated DNs; user needs to be in ANY of them

Example: `CN=StatpingUsers,OU=Groups,DC=example,DC=com`

## User Provisioning Flow

1. User authenticates via LDAP
2. If Authorized Group is enabled, group membership is verified
3. If user doesn't exist locally, they are auto-provisioned with:
   - Username from LDAP
   - Email from LDAP
   - Random 32-character password
   - **Disabled state** (`enabled=false`)
4. User cannot log in until an administrator enables their account

## Role Management

**Important**: All role assignment (admin/user) is managed within Statping, not via LDAP groups.

- New LDAP users are created with `admin=false`
- New LDAP users are created with `enabled=false` (pending approval)
- An administrator must:
  1. Enable the user account
  2. Optionally grant admin privileges

This ensures full control over who can access and administer the application.

## Database Schema

### Core Table (LDAP Settings)

| Column | Type | Description |
|--------|------|-------------|
| `ldap_enabled` | BOOL | Enable LDAP authentication |
| `ldap_host` | VARCHAR | LDAP server hostname |
| `ldap_port` | INT | LDAP server port (default: 636) |
| `ldap_start_tls` | BOOL | Use StartTLS (ports 389/3268) |
| `ldap_skip_verify` | BOOL | Skip TLS certificate verification |
| `ldap_bind_dn` | VARCHAR | Service account DN |
| `ldap_bind_password` | VARCHAR | Service account password |
| `ldap_base_dn` | VARCHAR | Search base DN |
| `ldap_user_filter` | VARCHAR | User search filter |
| `ldap_username_attr` | VARCHAR | Username attribute |
| `ldap_email_attr` | VARCHAR | Email attribute |
| `ldap_authorized_group_enabled` | BOOL | Require group membership |
| `ldap_authorized_group` | VARCHAR | Authorized group DN(s) |
| `ldap_template` | VARCHAR | Template name |

### Users Table

| Column | Type | Description |
|--------|------|-------------|
| `enabled` | BOOL | User account enabled (default: true for local, false for new LDAP) |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/ldap` | Get LDAP settings |
| POST | `/api/ldap` | Save LDAP settings |
| POST | `/api/ldap/test` | Test LDAP connection |
| GET | `/api/ldap/templates` | Get available templates |

## Security Considerations

1. **Use LDAPS (port 636/3269)** for production environments
2. **Do not skip TLS verification** in production
3. **Use a dedicated service account** with minimal permissions (read-only search)
4. **Enable Authorized Group** to restrict access to specific users
5. **Review pending users** regularly and enable only authorized accounts
