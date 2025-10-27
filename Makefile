DISK=$$HOME/disks
DATA=$(DISK)/fs/pb_data
BACKUP=$(DISK)/fs/pb_backup
STORAGE_DATA=$(DISK)/s3/

# Common setup tasks
setup-common:
	@test -d $(DISK) || mkdir -p $(DISK)
	mkdir -p $(DATA)
	sudo chown -R 1000:1000 $(DATA)
	@docker volume inspect pb_data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(DATA) pb_data

setup-backup:
	mkdir -p $(BACKUP)
	sudo chown -R 1000:1000 $(BACKUP)
	@docker volume inspect pb_backup >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(BACKUP) pb_backup

setup-storage:
	mkdir -p $(STORAGE_DATA)
	sudo chown -R 1000:1000 $(STORAGE_DATA)
	@docker volume inspect s3_data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(STORAGE_DATA) s3_data

# Common cleanup tasks
clean-common:
	docker volume rm pb_data >/dev/null 2>&1 || true
	sudo rm -rf $(DATA)

clean-backup:
	docker volume rm pb_backup >/dev/null 2>&1 || true
	sudo rm -rf $(BACKUP)

clean-storage:
	docker volume rm s3_data >/dev/null 2>&1 || true
	sudo rm -rf $(STORAGE_DATA)

# Setup for filesystem-based storage (local file replication)
setup-fs: setup-common setup-backup
	@echo "Filesystem setup complete. Use 'docker compose up' for local file replication."

# Setup for MinIO-based storage (S3-compatible replication)
setup-minio: setup-common setup-storage
	@echo "MinIO setup complete. Use 'docker compose -f docker-compose.minio.yml up' for S3-compatible storage."

# Setup for RustFS-based storage (high-performance S3-compatible replication)
setup-rustfs: setup-common setup-storage
	@echo "RustFS setup complete. Use 'docker compose -f docker-compose.rustfs.yml up' for high-performance S3-compatible storage."

# Setup for Garage-based storage (distributed S3-compatible replication)
setup-garage: setup-common setup-storage
	mkdir -p $(STORAGE_DATA)/garage/data
	mkdir -p $(STORAGE_DATA)/garage/meta
	sudo chown -R 1000:1000 $(STORAGE_DATA)/garage/data
	sudo chown -R 1000:1000 $(STORAGE_DATA)/garage/meta
	@docker volume inspect garage-data >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(STORAGE_DATA)/garage/data garage-data
	@docker volume inspect garage-meta >/dev/null 2>&1 || docker volume create --driver local --opt type=none --opt o=bind --opt device=$(STORAGE_DATA)/garage/meta garage-meta
	@echo "Garage setup complete. Use 'docker compose -f docker-compose.garage.yml up' for distributed S3-compatible storage."

# Clean filesystem setup
clean-fs: clean-common clean-backup
	@echo "Filesystem cleanup complete."

# Clean MinIO setup
clean-minio: clean-common clean-storage
	@echo "MinIO cleanup complete."

# Clean RustFS setup
clean-rustfs: clean-common clean-storage
	@echo "RustFS cleanup complete."

# Clean Garage setup
clean-garage: clean-common clean-storage
	docker volume rm garage-data >/dev/null 2>&1 || true
	docker volume rm garage-meta >/dev/null 2>&1 || true
	sudo rm -rf $(STORAGE_DATA)/garage/data
	sudo rm -rf $(STORAGE_DATA)/garage/meta
	@echo "Garage cleanup complete."

up-fs: setup-fs
	docker compose --env-file .env up --build -d

down-fs: 
	docker compose down -v
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