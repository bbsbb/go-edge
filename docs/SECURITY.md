<!-- last-reviewed: 2026-02-15 content-hash: 8c207d5f -->
# Security

Security model and practices.

## Multi-Tenancy: Row-Level Security (RLS)

### Model

All application tables live in the `app` schema. Tenant isolation is enforced at the database level using PostgreSQL Row-Level Security.

Each RLS-protected table has a policy:

```sql
CREATE POLICY organization_isolation_policy ON app.<table>
USING (organization_id = current_setting('app.current_organization')::UUID);
```

Enable RLS on tables that contain tenant-scoped data.

### How RLS Is Applied

1. **Organization middleware** extracts the organization slug from the request subdomain (via `X-Forwarded-Host` or `Host` header)
2. The organization is loaded and placed into `context.Context` via `domain.ContextWithOrganization()`
3. **`rlsfx.DB.Tx()`** starts a pgx transaction and executes `SET LOCAL app.current_organization = '<org-id>'`
4. All SQLC-generated queries within that transaction are automatically filtered by the RLS policy
5. `SET LOCAL` scoping ensures the variable is cleared when the transaction ends
6. **Nested transactions** use PostgreSQL savepoints. The inner `Tx()` saves the parent's RLS value, sets its own, and restores the parent's value after commit.

### Request Lifecycle: Organization Context Flow

The full multi-tenant request lifecycle from HTTP request to RLS-filtered query:

1. **HTTP request arrives** with tenant identifier in the `X-Organization-Slug` header (or subdomain via `X-Forwarded-Host`)
2. **`WithOrganization()` middleware** (`core/transport/http/middleware`) extracts the slug. Routes listed in `SkipPaths` (e.g., `/healthz`, `/readyz`) bypass this middleware.
3. **Organization lookup** — the middleware calls an `OrganizationLoader` (implemented by the app's persistence layer) using `*pgxpool.Pool` directly (outside RLS, since this table establishes tenant context)
4. **Organization stored in context** — `domain.ContextWithOrganization()` places the organization in `context.Context`, making it available to all downstream layers
5. **Handler receives request** — extracts path params, binds request body, calls the appropriate service method
6. **Service calls repository** — the repository depends on `*rlsfx.DB` for tenant-scoped tables
7. **`rlsfx.DB.Tx()` starts a transaction** — reads the organization from context, executes `SET LOCAL app.current_organization = '<org-id>'` to activate the RLS policy for this transaction
8. **SQLC-generated queries execute** — PostgreSQL's RLS policy automatically filters rows by `organization_id = current_setting('app.current_organization')::UUID`
9. **Transaction commits** — `SET LOCAL` scoping clears the session variable automatically

This flow ensures that tenant isolation is enforced at the database level, not in application code. A missing or incorrect organization in context causes `rlsfx.DB.Tx()` to fail before any query executes.

### Important Constraints

- RLS is only active inside `rlsfx.DB.Tx()` transactions. Non-RLS repos use `*pgxpool.Pool` directly.
- If `Tx()` is called without an organization in context, it returns an error without starting a transaction.
- The dependency type declares intent: `*rlsfx.DB` = RLS-enforced, `*pgxpool.Pool` = direct access.
- **DO NOT** use `*pgxpool.Pool` to query RLS-protected tables — this bypasses tenant isolation entirely. Choose the dependency type based on whether the table has an RLS policy.
- Non-RLS repos participate in ambient transactions via `psqlfx.TxFromContext()` for transaction composability.
- Tables that *establish* tenant context (e.g., organizations) sit **outside** RLS. They are queried before any RLS transaction exists (middleware resolves slug → org). See [ARCHITECTURE.md — RLS boundary](../ARCHITECTURE.md#rls-boundary-what-sits-inside-vs-outside) for the full pattern.

### RLS Configuration

```go
type RLS struct {
    Schema string  // e.g., "app"
    Field  string  // e.g., "current_organization"
}
```

The schema and field combine to form the PostgreSQL session variable name: `app.current_organization`.

## Authentication

### Approach: Passwordless Login Codes + JWT

The framework supports passwordless authentication via login codes and JWT tokens. Implementation details:

**Login flow pattern:**

1. Client sends email to initiate login
2. Server generates a short alphanumeric code (using `crypto/rand`)
3. Code is stored with an expiry
4. Client submits email + code to verify
5. On success, server returns a signed JWT

**JWT tokens:**
- Signed with HMAC-SHA256 using `lestrrat-go/jwx/v3`
- Configurable subject, issuer, and expiry
- Signing key: minimum 32 characters, configured via secret store

### Auth Configuration

```yaml
auth:
  jwt_signing_key: "secret://jwt-signing-key"  # resolved via secret store
  jwt_expiry: "24h"
  code_expiry: "10m"
```

## Secret Management: `secret://` Pattern

### How It Works

Configuration values can reference secrets by using the `secret://` prefix:

```yaml
database:
  password: "secret://db-password"
auth:
  jwt_signing_key: "secret://jwt-signing-key"
```

Or via environment variables:

```bash
DATABASE_PASSWORD=secret://db-password
```

### Resolution

During configuration loading, the `secretLookuper` intercepts values with the `secret://` prefix:

1. Strip the `secret://` prefix to get the secret name
2. Call `secretstore.Service.GetSecretValue(secretName)` to resolve the actual value
3. Return the resolved value in place of the `secret://` reference

If no `secretstore.Service` is provided (e.g., in development), `secret://` values are not resolved and the raw string is used as-is.

### Secret Store Interface

```go
type Service interface {
    GetSecretValue(secretName string) (string, error)
}
```

The interface is intentionally minimal. Implementations can back onto AWS Secrets Manager, Vault, or any other secret backend. The configuration layer doesn't know or care which backend is used.

### Adapters

The `secretstore` package provides built-in adapters. Applications select the adapter per environment in their configuration wiring.

| Adapter | Package | Resolves from | Use case |
|---------|---------|---------------|----------|
| `EnvService` | `core/secretstore` | Environment variables | Development, testing, CI |

`EnvService` converts secret names to environment variable keys: hyphens become underscores, names are uppercased. An optional prefix scopes lookups (e.g., prefix `"SECRET"` maps `"db-password"` → `SECRET_DB_PASSWORD`).

Production adapters (AWS Secrets Manager, Vault, etc.) implement the same `Service` interface and are wired in the application's `AsFx()` method based on environment.

### Development vs Production

- **Development/Testing:** Use `secretstore.NewEnvService("")` with `secret://` references. Set secrets as environment variables.
- **Production:** Use a production adapter (e.g., AWS Secrets Manager) with the same `secret://` references.
