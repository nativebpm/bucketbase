# PocketBase File Storage with S3 Backup

A production-ready file storage solution where metadata is stored in SQLite, API is provided by PocketBase, and files are stored in S3-compatible storage or local filesystem with automatic backups via Litestream.

## Quick Start

### Development (Filesystem)
```bash
make up-fs
```

### Production Options

#### MinIO (Local S3-compatible)
```bash
make up-minio
```

#### RustFS (High-performance S3-compatible)
```bash
make up-rustfs
```

#### AWS S3 (Cloud storage)
```bash
make up-aws
```

- **PocketBase:** http://localhost:8090/_/#/login (admin@example.com / admin123)
- **MinIO Console:** http://localhost:9001 (credentials in `.env.minio`)
- **RustFS Console:** http://localhost:9001 (credentials in `.env.rustfs`)

## Setup

1. **Filesystem storage**: `make setup-fs` - creates pb_data and pb_backup volumes
2. **MinIO storage**: `make setup-minio` - creates pb_data and s3_data volumes
3. **RustFS storage**: `make setup-rustfs` - creates pb_data and s3_data volumes
4. **AWS S3 storage**: `make setup-aws` - creates pb_data volume only

## Development vs Production

### Development
Use the filesystem backend for local development:
```bash
make up-fs
```
- Fast local file operations
- No external dependencies
- Automatic local backups

### Production
Choose based on your requirements:

- **MinIO**: Self-hosted S3-compatible storage, good for on-premises deployments
- **RustFS**: High-performance alternative to MinIO, better for high-throughput workloads  
- **AWS S3**: Cloud storage with high availability and scalability

## Environment Files

- `.env` - for filesystem setup
- `.env.minio` - for MinIO setup
- `.env.rustfs` - for RustFS setup
- `.env.aws` - for AWS S3 setup

Generate encryption key: `openssl rand -hex 32`

## Storage Backends

| Backend | Type | Setup Command | Compose File | Environment File |
|---------|------|---------------|--------------|------------------|
| Filesystem | Local | `make setup-fs` | `docker-compose.yml` | `.env` |
| MinIO | Local S3 | `make setup-minio` | `docker-compose.minio.yml` | `.env.minio` |
| RustFS | Local S3 | `make setup-rustfs` | `docker-compose.rustfs.yml` | `.env.rustfs` |
| AWS S3 | Cloud | `make setup-aws` | `docker-compose.aws.yml` | `.env.aws` |

## Switching Between Backends

To switch between different storage backends:

1. Stop the current setup: `make down-<current-backend>`
2. Set up the new backend: `make setup-<new-backend>`
3. Start with the new backend: `make up-<new-backend>`

Example:
```bash
make down-minio
make setup-rustfs
make up-rustfs
```

## Troubleshooting

### Common Issues

- **Permission denied**: Ensure Docker has access to the disk directories. You may need to run `sudo chown -R 1000:1000 $HOME/disks`
- **Port already in use**: Check if ports 8090, 9000, 9001 are available
- **Volume conflicts**: Run `make clean-<backend>` to remove old volumes before switching backends
- **AWS credentials**: For AWS S3, ensure your `.env.aws` file has valid AWS credentials with S3 permissions
- **Environment variables not set**: The Makefile automatically loads the appropriate `.env` file. If you run docker-compose directly, use `--env-file .env.<backend>` flag

### Health Checks

All services include health checks. Monitor container status with:
```bash
docker ps
docker logs <container-name>
```

## Sync Interval Costs

https://litestream.io/reference/config/#sync-interval-costs

## Features

- ✅ Metadata storage in SQLite database
- ✅ RESTful API provided by PocketBase
- ✅ File storage in multiple backends (Filesystem, MinIO, RustFS, AWS S3)
- ✅ SQLite database with Litestream replication
- ✅ Automatic backups to S3-compatible storage
- ✅ Docker Compose with health checks
- ✅ External volumes for persistence
- ✅ Automatic superuser creation
- ✅ Production resource limits
- ✅ Multiple storage backend support