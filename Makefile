DISK=$$HOME/disks

# Root paths for cleanup
FS_ROOT=$(DISK)/fs
MINIO_ROOT=$(DISK)/minio
RUSTFS_ROOT=$(DISK)/rustfs
GARAGE_ROOT=$(DISK)/garage
SEAWEEDFS_ROOT=$(DISK)/seaweedfs
RCLONE_ROOT=$(DISK)/rclone

# Pocketbase setup tasks
setup-pocketbase:
	mkdir -p $(DISK)/fs/pb_data

# Pocketbase cleanup tasks
clean-pocketbase:
	rm -rf $(DISK)/fs/pb_data

# Setup for filesystem-based storage (local file replication)
setup-fs: setup-pocketbase
	mkdir -p $(DISK)/fs/pb_backup

# Setup for MinIO-based storage (S3-compatible replication)
setup-minio: setup-pocketbase
	mkdir -p $(DISK)/minio/

# Setup for RustFS-based storage (high-performance S3-compatible replication)
setup-rustfs: setup-pocketbase
	mkdir -p $(DISK)/rustfs/

# Setup for Garage-based storage (distributed S3-compatible replication)
setup-garage: setup-pocketbase
	mkdir -p $(DISK)/garage/data
	mkdir -p $(DISK)/garage/meta

# Setup for SeaweedFS-based storage (S3-compatible replication)
setup-seaweedfs: setup-pocketbase
	mkdir -p $(DISK)/seaweedfs/

# Setup for rclone-based storage (S3-compatible replication)
setup-rclone: setup-pocketbase
	mkdir -p $(DISK)/rclone/

# Clean filesystem setup
clean-fs:
	rm -rf $(FS_ROOT)

# Clean MinIO setup
clean-minio:
	rm -rf $(MINIO_ROOT)

# Clean RustFS setup
clean-rustfs:
	rm -rf $(RUSTFS_ROOT)

# Clean Garage setup
clean-garage:
	rm -rf $(GARAGE_ROOT)

# Clean SeaweedFS setup
clean-seaweedfs:
	rm -rf $(SEAWEEDFS_ROOT)

# Clean rclone setup
clean-rclone:
	rm -rf $(RCLONE_ROOT)

up-seaweedfs: setup-seaweedfs
	USER_UID=$(shell id -u) USER_GID=$(shell id -g) docker compose --env-file .env.seaweedfs -f docker-compose.seaweedfs.yml up --build -d

down-seaweedfs:
	docker compose -f docker-compose.seaweedfs.yml down -v
	make clean-seaweedfs
	make clean-pocketbase

up-fs: setup-fs
	USER_UID=$(shell id -u) USER_GID=$(shell id -g) docker compose --env-file .env -f docker-compose.yml up --build -d

down-fs: 
	docker compose -f docker-compose.yml down -v
	make clean-fs
	make clean-pocketbase

up-minio: setup-minio
	USER_UID=$(shell id -u) USER_GID=$(shell id -g) docker compose --env-file .env.minio -f docker-compose.minio.yml up --build -d

down-minio:
	docker compose -f docker-compose.minio.yml down -v
	make clean-minio
	make clean-pocketbase

up-rustfs: setup-rustfs
	USER_UID=$(shell id -u) USER_GID=$(shell id -g) docker compose --env-file .env.rustfs -f docker-compose.rustfs.yml up --build -d

down-rustfs:
	docker compose -f docker-compose.rustfs.yml down -v
	make clean-rustfs
	make clean-pocketbase

up-garage: setup-garage
	USER_UID=$(shell id -u) USER_GID=$(shell id -g) docker compose --env-file .env.garage -f docker-compose.garage.yml up --build -d

down-garage:
	docker compose -f docker-compose.garage.yml down -v
	make clean-garage
	make clean-pocketbase

up-rclone: setup-rclone
	USER_UID=$(shell id -u) USER_GID=$(shell id -g) docker compose --env-file .env.rclone -f docker-compose.rclone.yml up --build -d

down-rclone:
	docker compose -f docker-compose.rclone.yml down -v
	make clean-rclone
	make clean-pocketbase

up-aws: setup-pocketbase
	USER_UID=$(shell id -u) USER_GID=$(shell id -g) docker compose --env-file .env.aws -f docker-compose.aws.yml up --build -d

down-aws:
	docker compose -f docker-compose.aws.yml down -v
	make clean-pocketbase
