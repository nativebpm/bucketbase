# PocketStream - Admin UI + REST API for any S3 Storage
[github.com/nativebpm/pocketstream](https://github.com/nativebpm/pocketstream)

File storage with SQLite metadata, PocketBase API, S3/local storage, and instant replication backups via Litestream.

## Key Features

- Simple technology stack: Docker + SQLite + Litestream + PocketBase + S3 backends
- Shell-script-free, Go binaries for better security and portability
- Simple deployment with Docker and Makefile
- Automatic backups with Litestream for seamless recovery
- Automatic database recovery with Litestream restore
- Built-in REST API and Admin UI via PocketBase
- High reliability with minimal maintenance requirements

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


## Useful Links

- [PocketBase Production Guide](https://pocketbase.io/docs/going-to-production/)
- [Litestream Docker Guide](https://litestream.io/guides/docker/#running-in-the-same-container)
- [Litestream Sync Interval Costs](https://litestream.io/reference/config/#sync-interval-costs)
- [Litestream Configuration Reference](https://litestream.io/reference/config/)
- [Litestream Restore Reference](https://litestream.io/reference/restore/)

## Requirements

- Docker
- Make
- Go 1.24+ (for building from source)

## Setup

- **Filesystem**: Fast local file operations, no external dependencies, automatic local backups
```bash
make up-fs
```

- **MinIO**: Self-hosted S3-compatible storage, suitable for on-premises deployments. Uses MinIO installed from source, latest release with CVE fixes (https://github.com/minio/minio/releases)
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

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
