#!/usr/bin/env bash
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.gdrive.yml"
TMP_DIR="/tmp/pocketstream-gdrive"
PB_URL="http://localhost:8090"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="admin12345678"

echo "=========================================================="
echo "🚀 Starting PocketStream End-to-End (E2E) Test Suite"
echo "=========================================================="

cleanup() {
    echo ""
    echo "🧹 Cleaning up test resources..."
    docker-compose --env-file "${ROOT_DIR}/.env.e2e" -f "${COMPOSE_FILE}" down -v --remove-orphans >/dev/null 2>&1 || docker-compose -f "${COMPOSE_FILE}" down -v --remove-orphans >/dev/null 2>&1 || true
    rm -f "${ROOT_DIR}/.env.e2e"
    rm -f /tmp/e2e_upload_test.txt
    rm -f /tmp/e2e_downloaded_test.txt
    echo "✅ Cleanup complete"
}

trap cleanup EXIT

# 0. Prepare E2E environment
echo "📝 Step 0: Generating .env.e2e configuration..."
cat <<EOF > "${ROOT_DIR}/.env.e2e"
POCKETBASE_ADMIN_EMAIL=${ADMIN_EMAIL}
POCKETBASE_ADMIN_PASSWORD=${ADMIN_PASSWORD}
PROFILE=docker
S3_ENABLED=true
S3_ENDPOINT=http://rclone:8334
S3_ACCESS_KEY=admin
S3_SECRET_KEY=admin123
S3_BUCKET=pocketstream
LITESTREAM_REPLICA_TYPE=file
LITESTREAM_DB_PATH=/pb_data/data.db
LITESTREAM_BACKUP_PATH=/pb_backup
LITESTREAM_SYNC_INTERVAL=1s
LITESTREAM_CHECKPOINT_INTERVAL=1s
EOF

# 1. Start stack cleanly
echo "📦 Step 1: Starting clean Docker Compose stack..."
docker-compose -f "${COMPOSE_FILE}" down -v >/dev/null 2>&1 || true
rm -rf "${TMP_DIR}/pb_data" "${TMP_DIR}/pb_backup"
mkdir -p "${TMP_DIR}/pb_data" "${TMP_DIR}/pb_backup"

docker-compose --env-file "${ROOT_DIR}/.env.e2e" -f "${COMPOSE_FILE}" up -d

echo "⏳ Waiting for PocketBase and S3 healthcheck..."
MAX_RETRIES=30
COUNT=0
until curl -s -f "${PB_URL}/api/health" >/dev/null 2>&1; do
    sleep 1
    COUNT=$((COUNT+1))
    if [ "${COUNT}" -ge "${MAX_RETRIES}" ]; then
        echo "❌ Timeout waiting for PocketBase to become healthy!"
        docker-compose -f "${COMPOSE_FILE}" logs
        exit 1
    fi
done

HEALTH_RESP=$(curl -s "${PB_URL}/api/health")
echo "✅ PocketBase is healthy: ${HEALTH_RESP}"

# 2. Superuser Authentication
echo ""
echo "🔑 Step 2: Authenticating superuser (${ADMIN_EMAIL})..."
AUTH_RESP=$(curl -s -X POST "${PB_URL}/api/collections/_superusers/auth-with-password" \
    -H "Content-Type: application/json" \
    -d "{\"identity\": \"${ADMIN_EMAIL}\", \"password\": \"${ADMIN_PASSWORD}\"}")

AUTH_TOKEN=$(echo "${AUTH_RESP}" | jq -r '.token // empty')
if [ -z "${AUTH_TOKEN}" ]; then
    echo "❌ Failed to authenticate superuser! Response: ${AUTH_RESP}"
    exit 1
fi
echo "✅ Authenticated successfully! JWT token received (length: ${#AUTH_TOKEN})"

# 3. Create Collection
echo ""
echo "📁 Step 3: Creating collection 'e2e_documents'..."
COLLECTION_PAYLOAD='{
    "name": "e2e_documents",
    "type": "base",
    "viewRule": "",
    "listRule": "",
    "createRule": "",
    "updateRule": "",
    "deleteRule": "",
    "fields": [
        {
            "name": "title",
            "type": "text",
            "required": true
        },
        {
            "name": "attachment",
            "type": "file",
            "maxSelect": 1,
            "maxSize": 10485760
        }
    ]
}'

CREATE_COLL_RESP=$(curl -s -X POST "${PB_URL}/api/collections" \
    -H "Authorization: ${AUTH_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "${COLLECTION_PAYLOAD}")

COLL_ID=$(echo "${CREATE_COLL_RESP}" | jq -r '.id // empty')
if [ -z "${COLL_ID}" ]; then
    echo "❌ Failed to create collection! Response: ${CREATE_COLL_RESP}"
    exit 1
fi
echo "✅ Collection 'e2e_documents' created with ID: ${COLL_ID}"

# 4. Upload Record with File Attachment
echo ""
echo "📄 Step 4: Creating record with file attachment..."
TEST_CONTENT="PocketStream End-to-End Test Payload at $(date) - Random: $RANDOM"
echo "${TEST_CONTENT}" > /tmp/e2e_upload_test.txt
ORIGINAL_HASH=$(shasum -a 256 /tmp/e2e_upload_test.txt | awk '{print $1}')
echo "Generated test file SHA256: ${ORIGINAL_HASH}"

CREATE_RECORD_RESP=$(curl -s -X POST "${PB_URL}/api/collections/e2e_documents/records" \
    -H "Authorization: ${AUTH_TOKEN}" \
    -F "title=E2E Test Document #1" \
    -F "attachment=@/tmp/e2e_upload_test.txt")

RECORD_ID=$(echo "${CREATE_RECORD_RESP}" | jq -r '.id // empty')
ATTACHMENT_NAME=$(echo "${CREATE_RECORD_RESP}" | jq -r '.attachment // empty')

if [ -z "${RECORD_ID}" ] || [ -z "${ATTACHMENT_NAME}" ]; then
    echo "❌ Failed to create record! Response: ${CREATE_RECORD_RESP}"
    exit 1
fi
echo "✅ Record created: ID=${RECORD_ID}, Attachment=${ATTACHMENT_NAME}"

# 5. Verify File Download Integrity (with polling for S3 eventual consistency)
echo ""
echo "🔍 Step 5: Verifying file download integrity from PocketBase..."
FILE_URL="${PB_URL}/api/files/e2e_documents/${RECORD_ID}/${ATTACHMENT_NAME}"
echo "Requesting: ${FILE_URL}"

MAX_DOWNLOAD_RETRIES=20
COUNT=0
HTTP_CODE=0
while [ "${COUNT}" -lt "${MAX_DOWNLOAD_RETRIES}" ]; do
    HTTP_CODE=$(curl -s -w "%{http_code}" -H "Authorization: ${AUTH_TOKEN}" "${FILE_URL}" -o /tmp/e2e_downloaded_test.txt)
    if [ "${HTTP_CODE}" = "200" ]; then
        break
    fi
    sleep 1
    COUNT=$((COUNT+1))
done
echo "File download HTTP status: ${HTTP_CODE} (after ${COUNT} retries)"

if [ "${HTTP_CODE}" != "200" ]; then
    echo "❌ Download failed with HTTP ${HTTP_CODE}!"
    echo "Response body: $(cat /tmp/e2e_downloaded_test.txt)"
    docker-compose --env-file "${ROOT_DIR}/.env.e2e" -f "${COMPOSE_FILE}" logs pocketbase
    exit 1
fi

DOWNLOADED_HASH=$(shasum -a 256 /tmp/e2e_downloaded_test.txt | awk '{print $1}')
echo "Downloaded file SHA256: ${DOWNLOADED_HASH}"

if [ "${ORIGINAL_HASH}" != "${DOWNLOADED_HASH}" ]; then
    echo "❌ Checksum mismatch! Expected ${ORIGINAL_HASH}, got ${DOWNLOADED_HASH}"
    exit 1
fi
echo "✅ File integrity verified: SHA256 checksums match perfectly!"

# 6. Verify Litestream Replication
echo ""
echo "🔄 Step 6: Verifying continuous Litestream WAL replication..."
sleep 3

LITESTREAM_LOGS=$(docker-compose --env-file "${ROOT_DIR}/.env.e2e" -f "${COMPOSE_FILE}" logs litestream 2>&1)
if echo "${LITESTREAM_LOGS}" | grep -qi "monitor error"; then
    echo "❌ Litestream monitor errors detected in logs!"
    echo "${LITESTREAM_LOGS}"
    exit 1
fi

if ! echo "${LITESTREAM_LOGS}" | grep -q "compaction complete"; then
    echo "⚠️ Waiting an additional 2 seconds for Litestream compaction..."
    sleep 2
    LITESTREAM_LOGS=$(docker-compose --env-file "${ROOT_DIR}/.env.e2e" -f "${COMPOSE_FILE}" logs litestream 2>&1)
fi

echo "✅ Litestream replication confirmed active without errors"

# 7. Disaster Recovery: Simulate Complete Database Loss
echo ""
echo "💥 Step 7: DISASTER RECOVERY TEST - Simulating complete database loss..."
echo "Stopping PocketBase..."
docker-compose --env-file "${ROOT_DIR}/.env.e2e" -f "${COMPOSE_FILE}" stop pocketbase >/dev/null

echo "Simulating disk failure: removing ${TMP_DIR}/pb_data/data.db and WAL files..."
rm -f "${TMP_DIR}/pb_data/data.db" "${TMP_DIR}/pb_data/data.db-wal" "${TMP_DIR}/pb_data/data.db-shm"
if [ -f "${TMP_DIR}/pb_data/data.db" ]; then
    echo "❌ Failed to delete data.db!"
    exit 1
fi
echo "Database file deleted: $(ls -la "${TMP_DIR}/pb_data")"

echo "Starting PocketBase (triggering automatic litestream restore)..."
docker-compose --env-file "${ROOT_DIR}/.env.e2e" -f "${COMPOSE_FILE}" start pocketbase >/dev/null

echo "Waiting for PocketBase to restore database and start..."
COUNT=0
until curl -s -f "${PB_URL}/api/health" >/dev/null 2>&1; do
    sleep 1
    COUNT=$((COUNT+1))
    if [ "${COUNT}" -ge "${MAX_RETRIES}" ]; then
        echo "❌ PocketBase failed to recover after database loss!"
        docker-compose --env-file "${ROOT_DIR}/.env.e2e" -f "${COMPOSE_FILE}" logs pocketbase
        exit 1
    fi
done
echo "✅ PocketBase recovered and healthy after disaster recovery!"

# 8. Verify Data After Disaster Recovery
echo ""
echo "✨ Step 8: Verifying data integrity post-recovery..."
# Re-authenticate to ensure token is fresh
AUTH_RESP=$(curl -s -X POST "${PB_URL}/api/collections/_superusers/auth-with-password" \
    -H "Content-Type: application/json" \
    -d "{\"identity\": \"${ADMIN_EMAIL}\", \"password\": \"${ADMIN_PASSWORD}\"}")
AUTH_TOKEN=$(echo "${AUTH_RESP}" | jq -r '.token // empty')

VERIFY_RECORD_RESP=$(curl -s "${PB_URL}/api/collections/e2e_documents/records/${RECORD_ID}" \
    -H "Authorization: ${AUTH_TOKEN}")

RECOVERED_TITLE=$(echo "${VERIFY_RECORD_RESP}" | jq -r '.title // empty')
RECOVERED_ATTACHMENT=$(echo "${VERIFY_RECORD_RESP}" | jq -r '.attachment // empty')

if [ "${RECOVERED_TITLE}" != "E2E Test Document #1" ]; then
    echo "❌ Recovered record mismatch! Got title: ${RECOVERED_TITLE}"
    exit 1
fi
echo "✅ Record successfully recovered from Litestream backup: ${RECOVERED_TITLE}"

# Download file from recovered instance with retry for S3 consistency
rm -f /tmp/e2e_downloaded_test.txt
COUNT=0
HTTP_CODE=0
while [ "${COUNT}" -lt "${MAX_DOWNLOAD_RETRIES}" ]; do
    HTTP_CODE=$(curl -s -w "%{http_code}" -H "Authorization: ${AUTH_TOKEN}" "${FILE_URL}" -o /tmp/e2e_downloaded_test.txt)
    if [ "${HTTP_CODE}" = "200" ]; then
        break
    fi
    sleep 1
    COUNT=$((COUNT+1))
done

if [ "${HTTP_CODE}" != "200" ]; then
    echo "❌ Post-recovery download failed with HTTP ${HTTP_CODE}!"
    exit 1
fi
RECOVERED_HASH=$(shasum -a 256 /tmp/e2e_downloaded_test.txt | awk '{print $1}')

if [ "${ORIGINAL_HASH}" != "${RECOVERED_HASH}" ]; then
    echo "❌ Post-recovery file checksum mismatch! Expected ${ORIGINAL_HASH}, got ${RECOVERED_HASH}"
    exit 1
fi
echo "✅ Post-recovery file integrity verified: SHA256 checksum matches original!"

echo ""
echo "=========================================================="
echo "🎉 ALL END-TO-END (E2E) TESTS PASSED SUCCESSFULLY! (4/4)"
echo "=========================================================="
