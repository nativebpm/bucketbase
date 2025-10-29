DISK=$$HOME/disks

# Filesystem-specific paths
FS_DATA=$(DISK)/fs/pb_data
FS_BACKUP=$(DISK)/fs/pb_backup

# Storage-specific paths
MINIO_DATA=$(DISK)/minio/
RUSTFS_DATA=$(DISK)/rustfs/
GARAGE_DATA=$(DISK)/garage/data
GARAGE_META=$(DISK)/garage/meta
SEAWEEDFS_DATA=$(DISK)/seaweedfs/
RCLONE_DATA=$(DISK)/rclone/

# Pocketbase setup tasks
setup-pocketbase:
	mkdir -p $(FS_DATA)
	sudo chown -R 1000:1000 $(FS_DATA)
	@docker volume inspect pb_data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(FS_DATA) pb_data

# Pocketbase cleanup tasks
clean-pocketbase:
	docker volume rm pb_data >/dev/null 2>&1 || true
	sudo rm -rf $(FS_DATA)

# Setup for filesystem-based storage (local file replication)
setup-fs: setup-pocketbase
	mkdir -p $(FS_BACKUP)
	sudo chown -R 1000:1000 $(FS_BACKUP)
	@docker volume inspect pb_backup >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(FS_BACKUP) pb_backup

# Setup for MinIO-based storage (S3-compatible replication)
setup-minio: setup-pocketbase
	mkdir -p $(MINIO_DATA)
	sudo chown -R 1000:1000 $(MINIO_DATA)
	@docker volume inspect minio-data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(MINIO_DATA) minio-data

# Setup for RustFS-based storage (high-performance S3-compatible replication)
setup-rustfs: setup-pocketbase
	mkdir -p $(RUSTFS_DATA)
	sudo chown -R 1000:1000 $(RUSTFS_DATA)
	@docker volume inspect rustfs-data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(RUSTFS_DATA) rustfs-data

# Setup for Garage-based storage (distributed S3-compatible replication)
setup-garage: setup-pocketbase
	mkdir -p $(GARAGE_DATA)
	mkdir -p $(GARAGE_META)
	sudo chown -R 1000:1000 $(GARAGE_DATA)
	sudo chown -R 1000:1000 $(GARAGE_META)
	@docker volume inspect garage-data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(GARAGE_DATA) garage-data
	@docker volume inspect garage-meta >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(GARAGE_META) garage-meta

# Setup for SeaweedFS-based storage (S3-compatible replication)
setup-seaweedfs: setup-pocketbase
	mkdir -p $(SEAWEEDFS_DATA)
	sudo chown -R 1000:1000 $(SEAWEEDFS_DATA)
	@docker volume inspect seaweedfs-data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(SEAWEEDFS_DATA) seaweedfs-data

# Setup for rclone-based storage (S3-compatible replication)
setup-rclone: setup-pocketbase
	mkdir -p $(RCLONE_DATA)
	sudo chown -R 1000:1000 $(RCLONE_DATA)
	@docker volume inspect rclone-data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(RCLONE_DATA) rclone-data

# Clean filesystem setup
clean-fs: clean-pocketbase
	docker volume rm pb_backup >/dev/null 2>&1 || true
	sudo rm -rf $(FS_BACKUP)

# Clean MinIO setup
clean-minio: clean-pocketbase
	docker volume rm minio-data >/dev/null 2>&1 || true
	sudo rm -rf $(MINIO_DATA)

# Clean RustFS setup
clean-rustfs: clean-pocketbase
	docker volume rm rustfs-data >/dev/null 2>&1 || true
	sudo rm -rf $(RUSTFS_DATA)

# Clean Garage setup
clean-garage: clean-pocketbase
	docker volume rm garage-data >/dev/null 2>&1 || true
	docker volume rm garage-meta >/dev/null 2>&1 || true
	sudo rm -rf $(GARAGE_DATA)
	sudo rm -rf $(GARAGE_META)

# Clean SeaweedFS setup
clean-seaweedfs: clean-pocketbase
	docker volume rm seaweedfs-data >/dev/null 2>&1 || true
	sudo rm -rf $(SEAWEEDFS_DATA)

# Clean rclone setup
clean-rclone: clean-pocketbase
	docker volume rm rclone-data >/dev/null 2>&1 || true
	sudo rm -rf $(RCLONE_DATA)

up-seaweedfs: setup-seaweedfs
	docker compose --env-file .env.seaweedfs -f docker-compose.seaweedfs.yml up --build -d

down-seaweedfs:
	docker compose -f docker-compose.seaweedfs.yml down -v
	make clean-seaweedfs

up-fs: setup-fs
	docker compose --env-file .env -f docker-compose.yml up --build -d

down-fs: 
	docker compose -f docker-compose.yml down -v
	make clean-fs

up-minio: setup-minio
	docker compose --env-file .env.minio -f docker-compose.minio.yml up --build -d

down-minio:
	docker compose -f docker-compose.minio.yml down -v
	make clean-minio

up-rustfs: setup-rustfs
	docker compose --env-file .env.rustfs -f docker-compose.rustfs.yml up --build -d

down-rustfs:
	docker compose -f docker-compose.rustfs.yml down -v
	make clean-rustfs

up-garage: setup-garage
	docker compose --env-file .env.garage -f docker-compose.garage.yml up --build -d

down-garage:
	docker compose -f docker-compose.garage.yml down -v
	make clean-garage

up-rclone: setup-rclone
	docker compose --env-file .env.rclone -f docker-compose.rclone.yml up --build -d

down-rclone:
	docker compose -f docker-compose.rclone.yml down -v
	make clean-rclone

up-aws: setup-pocketbase
	docker compose --env-file .env.aws -f docker-compose.aws.yml up --build -d

down-aws:
	docker compose -f docker-compose.aws.yml down -v
	make clean-aws