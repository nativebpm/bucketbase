DISK=$$HOME/disks
DATA=$(DISK)/fs/pb_data
BACKUP=$(DISK)/fs/pb_backup
STORAGE_DATA=$(DISK)/s3/

# Setup for filesystem-based storage (local file replication)
setup-fs:
	@echo "Setting up filesystem-based storage..."
	@test -d $(DISK) || mkdir -p $(DISK)
	mkdir -p $(DATA) $(BACKUP)
	sudo chown -R 1000:1000 $(DATA) $(BACKUP)
	@docker volume inspect pb_data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(DATA) pb_data
	@docker volume inspect pb_backup >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(BACKUP) pb_backup
	@echo "Filesystem setup complete. Use 'docker compose up' for local file replication."

# Setup for MinIO-based storage (S3-compatible replication)
setup-minio:
	@echo "Setting up MinIO-based storage..."
	@test -d $(DISK) || mkdir -p $(DISK)
	mkdir -p $(DATA) $(STORAGE_DATA)
	sudo chown -R 1000:1000 $(DATA) $(STORAGE_DATA)
	@docker volume inspect pb_data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(DATA) pb_data
	@docker volume inspect s3_data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(STORAGE_DATA) s3_data
	@echo "MinIO setup complete. Use 'docker compose -f docker-compose.minio.yml up' for S3-compatible storage."

# Clean filesystem setup
clean-fs:
	@echo "Cleaning filesystem-based storage..."
	docker volume rm pb_data pb_backup >/dev/null 2>&1 || true
	sudo rm -rf $(DATA) $(BACKUP)
	@echo "Filesystem cleanup complete."

# Clean MinIO setup
clean-minio:
	@echo "Cleaning MinIO-based storage..."
	docker volume rm pb_data s3_data >/dev/null 2>&1 || true
	sudo rm -rf $(DATA) $(STORAGE_DATA)
	@echo "MinIO cleanup complete."

up-fs: setup-fs
	docker compose up --build

down-fs: 
	docker compose down -v
	make clean-fs

up-minio: setup-minio
	docker compose -f docker-compose.minio.yml up --build

down-minio:
	docker compose -f docker-compose.minio.yml down -v
	make clean-minio