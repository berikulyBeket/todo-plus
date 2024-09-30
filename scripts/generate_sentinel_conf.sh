#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

CONF_DIR="./config"
SENTINEL_CONF="$CONF_DIR/sentinel.conf"
SENTINEL_LOCAL_CONF="$CONF_DIR/sentinel_local.conf"

mkdir -p "$CONF_DIR"

# Write the configuration content to sentinel.conf (using container name)
cat <<EOF > "$SENTINEL_CONF"
sentinel resolve-hostnames yes
sentinel auth-pass $REDIS_MASTER_NAME $REDIS_MASTER_PASSWORD
sentinel monitor $REDIS_MASTER_NAME redis-master 6379 2
sentinel down-after-milliseconds $REDIS_MASTER_NAME 5000
sentinel failover-timeout $REDIS_MASTER_NAME 10000
sentinel parallel-syncs $REDIS_MASTER_NAME 1
EOF

echo "Sentinel configuration generated at $SENTINEL_CONF."

# Write the configuration content to sentinel_local.conf (using host IP)
cat <<EOF > "$SENTINEL_LOCAL_CONF"
sentinel auth-pass $REDIS_MASTER_NAME $REDIS_MASTER_PASSWORD
sentinel monitor $REDIS_MASTER_NAME 192.168.100.126 6379 2
sentinel down-after-milliseconds $REDIS_MASTER_NAME 5000
sentinel failover-timeout $REDIS_MASTER_NAME 10000
sentinel parallel-syncs $REDIS_MASTER_NAME 1
EOF

echo "Sentinel local configuration generated at $SENTINEL_LOCAL_CONF."
