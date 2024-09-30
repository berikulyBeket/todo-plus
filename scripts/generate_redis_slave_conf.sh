#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

CONF_DIR="./config"
REDIS_SLAVE_CONF="$CONF_DIR/redis_slave.conf"
REDIS_SLAVE_LOCAL_CONF="$CONF_DIR/redis_slave_local.conf"

mkdir -p "$CONF_DIR"

# Write the configuration content to redis_slave.conf (for Docker containers)
cat <<EOF > "$REDIS_SLAVE_CONF"
appendonly yes
replicaof redis-master 6379
masterauth $REDIS_MASTER_PASSWORD
EOF

echo "Redis slave configuration generated at $REDIS_SLAVE_CONF."

# Write the configuration content to redis_slave_local.conf (for local use with IP)
cat <<EOF > "$REDIS_SLAVE_LOCAL_CONF"
appendonly yes
replicaof redis-master 6379
masterauth $REDIS_MASTER_PASSWORD
replica-announce-ip 192.168.100.126
replica-announce-port 6380
EOF

echo "Redis local slave configuration generated at $REDIS_SLAVE_LOCAL_CONF."
