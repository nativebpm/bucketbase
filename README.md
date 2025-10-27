# Admin UI + REST API for any S3 storage

## [PocketBase File Storage with S3 Backup](https://github.com/nativebpm/pocketbase)

A production-ready file storage solution where metadata is stored in SQLite, API is provided by PocketBase, and files are stored in S3-compatible storage or local filesystem with automatic backups via Litestream.

## Features

- ✅ Metadata storage in SQLite database
- ✅ RESTful API provided by PocketBase
- ✅ File storage in multiple backends (Filesystem, MinIO, RustFS, Garage, AWS S3)
- ✅ SQLite database with Litestream replication
- ✅ Automatic backups to S3-compatible storage
- ✅ Automatic S3 bucket creation on startup
- ✅ Docker Compose with health checks
- ✅ External volumes for persistence
- ✅ Automatic superuser creation
- ✅ Production resource limits
- ✅ Multiple storage backend support

## Setup

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
- **Garage**: Distributed S3-compatible storage with built-in replication, ideal for multi-node clusters
- **SeaweedFS**: Fast and scalable distributed file system with S3 API, good for high-performance workloads
- **AWS S3**: Cloud storage with high availability and scalability

### Links

- **PocketBase:** http://localhost:8090/_/#/login (admin@example.com / admin123)
- **MinIO Console:** http://localhost:9001 (credentials in `.env.minio`)
- **RustFS Console:** http://localhost:9001 (credentials in `.env.rustfs`)
- **Garage Console:** http://localhost:3900 (S3 API endpoint)

### Environment Files

- `.env` - Filesystem storage
- `.env.minio` - MinIO S3 storage
- `.env.rustfs` - RustFS S3 storage
- `.env.garage` - Garage S3 storage
- `.env.seaweedfs` - SeaweedFS S3 storage
- `.env.aws` - AWS S3 storage

Generate encryption key: `openssl rand -hex 32`

## Storage Backends

| Backend | Type | Setup Command | Compose File | Environment File |
|---------|------|---------------|--------------|------------------|
| Filesystem | Local | `make setup-fs` | `docker-compose.yml` | `.env` |
| MinIO | Local S3 | `make setup-minio` | `docker-compose.minio.yml` | `.env.minio` |
| RustFS | Local S3 | `make setup-rustfs` | `docker-compose.rustfs.yml` | `.env.rustfs` |
| Garage | Distributed S3 | `make setup-garage` | `docker-compose.garage.yml` | `.env.garage` |
| SeaweedFS | Distributed S3 | `make setup-seaweedfs` | `docker-compose.seaweedfs.yml` | `.env.seaweedfs` |
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
- **Litestream path issues**: Ensure `LITESTREAM_PATH` does not start with `/` to avoid double slashes in S3 object keys (e.g., use `app/pb_data/data.db` instead of `/app/pb_data/data.db`)

## Sync Interval Costs

Configuration follows the official Litestream documentation: https://litestream.io/reference/config/#sync-interval-costs

The setup has been verified for compliance with Litestream's S3 guide: https://litestream.io/guides/s3/