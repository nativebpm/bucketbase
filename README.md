# Admin UI + REST API for any S3 storage

## [PocketBase File Storage with S3 Backup](https://github.com/nativebpm/pocketbase)

A production-ready file storage solution where metadata is stored in SQLite, API is provided by PocketBase, and files are stored in S3-compatible storage or local filesystem with automatic backups via Litestream.

## Storage Backends

| Backend    | Type           |
|------------|----------------|
| Filesystem | Local          |
| MinIO      | Local S3       |
| RustFS     | Local S3       |
| Garage     | Distributed S3 |
| SeaweedFS  | Distributed S3 |
| AWS S3     | Cloud          |

## Setup

- **Filesystem**: Fast local file operations, no external dependencies, automatic local backups
```bash
make up-fs
```
- **MinIO**: Self-hosted S3-compatible storage, good for on-premises deployments
```bash
make up-minio
```
- **RustFS**: High-performance alternative to MinIO, better for high-throughput workloads
```bash
make up-rustfs
```
- **Garage**: Distributed S3-compatible storage with built-in replication, ideal for multi-node clusters
```bash
make up-garage
```
- **SeaweedFS**: Fast and scalable distributed file system with S3 API, good for high-performance workloads
```bash
make up-seaweedfs
```
- **AWS S3**: Cloud storage with high availability and scalability
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