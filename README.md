## Admin UI + REST API for any S3 Storage

### BucketBase - File Storage with S3 Backup
[github.com/nativebpm/bucketbase](https://github.com/nativebpm/bucketbase)

A production-ready file storage solution where metadata is stored in SQLite, the API is provided by PocketBase, and files are stored in S3-compatible storage or the local filesystem, with automatic backups handled by Litestream.

This is a proven and practical technology stack: Docker + SQLite + Litestream + PocketBase (Go, Admin UI, REST API) + any S3-compatible backend (Amazon S3 SDK).

## Storage Backends

| Backend    | Type           |
|------------|----------------|
| Filesystem | Local          |
| MinIO      | Local S3       |
| RustFS     | Local S3       |
| Garage     | Distributed S3 |
| SeaweedFS  | Distributed S3 |
| rclone     | Local S3       |
| AWS S3     | Cloud          |

## Sync Interval Costs

The configuration follows the official Litestream documentation.
- [litestream.io/reference/config/#sync-interval-costs](https://litestream.io/reference/config/#sync-interval-costs)
- [litestream.io/guides/s3/](https://litestream.io/guides/s3/)

## Setup

- **Filesystem**: Fast local file operations, no external dependencies, automatic local backups
```bash
make up-fs
```

- **MinIO**: Self-hosted S3-compatible storage, suitable for on-premises deployments
```bash
make up-minio
```

- **RustFS**: High-performance alternative to MinIO, optimized for high-throughput workloads
```bash
make up-rustfs
```

- **Garage**: Distributed S3-compatible storage with built-in replication, ideal for multi-node clusters
```bash
make up-garage
```

- **SeaweedFS**: Fast and scalable distributed file system with an S3 API, suited for high-performance workloads
```bash
make up-seaweedfs
```

- **rclone**: Flexible S3-compatible layer using `rclone serve s3`, supports various storage backends
```bash
make up-rclone
```

- **AWS S3**: Cloud-based storage offering high availability and scalability
```bash
make up-aws
```

### Links

- **PocketBase:** http://localhost:8090/_/#/login (admin@example.com / admin123)
- **MinIO Console:** http://localhost:9001 (credentials in `.env.minio`)
- **RustFS Console:** http://localhost:9001 (credentials in `.env.rustfs`)
- **SeaweedFS Console:** http://localhost:9333 (credentials in `.env.seaweedfs`)
- **Garage Admin API:** http://localhost:3903 (token in `garage.toml`)

## Clean

```bash
make down-<backend>
```

## Production Notes

I tested the stack extensively in real deployment scenarios, including integration with MinIO as the S3-compatible backend. It has proven stable and reliable under load, with smooth recovery and backup using Litestream. PocketBase provides a solid REST API and admin UI out of the box, while SQLite and Docker make setup and deployment simple and consistent. With MinIO ensuring durability and backup, the stack has demonstrated production-grade reliability and minimal maintenance requirements.
