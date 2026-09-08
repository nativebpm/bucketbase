.PHONY: help up-fs down-fs up-minio down-minio up-rustfs down-rustfs up-garage down-garage up-seaweedfs down-seaweedfs up-rclone down-rclone up-aws down-aws up-gdrive down-gdrive

help:
	@echo "PocketStream Pure Docker Commands:"
	@echo "  make up-fs / down-fs          - Local filesystem replication"
	@echo "  make up-minio / down-minio    - MinIO S3 storage"
	@echo "  make up-rustfs / down-rustfs  - RustFS S3 storage"
	@echo "  make up-garage / down-garage  - Garage S3 storage"
	@echo "  make up-seaweedfs / down-seaweedfs - SeaweedFS storage"
	@echo "  make up-rclone / down-rclone  - Rclone S3 storage"
	@echo "  make up-aws / down-aws        - AWS S3 cloud storage"
	@echo "  make up-gdrive / down-gdrive  - Google Drive via Rclone"

up-fs:
	docker compose --env-file .env -f docker-compose.yml up -d

down-fs:
	docker compose -f docker-compose.yml down -v

up-minio:
	docker compose --env-file .env.minio -f docker-compose.minio.yml up -d

down-minio:
	docker compose -f docker-compose.minio.yml down -v

up-rustfs:
	docker compose --env-file .env.rustfs -f docker-compose.rustfs.yml up -d

down-rustfs:
	docker compose -f docker-compose.rustfs.yml down -v

up-garage:
	docker compose --env-file .env.garage -f docker-compose.garage.yml up -d

down-garage:
	docker compose -f docker-compose.garage.yml down -v

up-seaweedfs:
	docker compose --env-file .env.seaweedfs -f docker-compose.seaweedfs.yml up -d

down-seaweedfs:
	docker compose -f docker-compose.seaweedfs.yml down -v

up-rclone:
	docker compose --env-file .env.rclone -f docker-compose.rclone.yml up -d

down-rclone:
	docker compose -f docker-compose.rclone.yml down -v

up-aws:
	docker compose --env-file .env.aws -f docker-compose.aws.yml up -d

down-aws:
	docker compose -f docker-compose.aws.yml down -v

up-gdrive:
	docker compose --env-file .env.gdrive -f docker-compose.gdrive.yml up -d

down-gdrive:
	docker compose -f docker-compose.gdrive.yml down -v
