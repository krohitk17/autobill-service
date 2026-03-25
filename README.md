# autobill-service
Backend service for AutoBill.

## Run With Docker

### 1. Build and start API + Postgres

```bash
docker compose up --build
```

The API will be available on `http://localhost:8001`.

This flow now includes:
- Postgres init script to create application DB/user and grants.
- SQL migrations are applied by Postgres init using scripts in `docker/postgres/migrations`.

### 2. Stop services

```bash
docker compose down
```

To also remove database data volume:

```bash
docker compose down -v
```

Note: Compose is self-contained and does not load env files. Use `.env` for local non-Docker testing only.

## PostgreSQL Strategy

### Local development

- Use the `postgres` service in `docker-compose.yml`.
- Data is persisted in the named volume `postgres_data`.
- The API connects using `DATABASE_HOST=postgres` (service name on Docker network).
- DB schema is created from SQL migration scripts under `docker/postgres/migrations`.

### Production deployment

- Prefer a managed PostgreSQL service (RDS, Cloud SQL, etc.) instead of running DB in the same host/container stack.
- Run API container separately and point env vars to managed DB host.
- Set `DATABASE_SSL_MODE=require` (or stricter, depending on provider).
- Keep automated backups and point-in-time recovery enabled.

### If you still self-host Postgres

- Keep database storage on persistent volumes/disks.
- Add regular backups (e.g., `pg_dump` cron + offsite storage).
- Restrict network access to Postgres; do not expose it publicly unless required.

## SQL Migrations

- Database/bootstrap script: `docker/postgres/initdb/01-create-app-db-and-user.sh`
- SQL schema migrations: `docker/postgres/migrations/*.sql`
- Setup verification script: `scripts/db/check-db-setup.sh`

Verify DB setup (role, database, grants, core tables):

```bash
sh scripts/db/check-db-setup.sh .env
```
