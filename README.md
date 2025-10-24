# PocketBase File Storage with S3 Backup

A production-ready file storage solution where metadata is stored in SQLite, API is provided by PocketBase, and files are stored in S3-compatible storage or local filesystem with automatic backups via Litestream.

## Quick Start

### Development (Filesystem)
```bash
make up-fs
```

### Production (MinIO)
```bash
make up-minio
```

PocketBase: http://localhost:8090  
MinIO Console: http://localhost:9001 (minioadmin/minioadmin)

**Default admin credentials:** admin@example.com / admin123 (created automatically)

## Setup

1. **Filesystem storage**: `make setup-fs` - creates pb_data and pb_backup volumes
2. **MinIO storage**: `make setup-minio` - creates pb_data and minio_data volumes

## Environment

- `.env` - for filesystem setup
- `.env.minio` - for MinIO setup

Generate encryption key: `openssl rand -hex 32`

## Features

- ✅ Metadata storage in SQLite database
- ✅ RESTful API provided by PocketBase
- ✅ File storage in S3-compatible storage or local filesystem
- ✅ SQLite database with Litestream replication
- ✅ Automatic backups to S3-compatible storage (MinIO, Amazon S3 etc)
- ✅ Docker Compose with health checks
- ✅ External volumes for persistence
- ✅ Automatic superuser creation
- ✅ Production resource limits