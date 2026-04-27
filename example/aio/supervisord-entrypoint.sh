#!/bin/bash
# Supervisor entrypoint - creates required directories before starting

set -e

# Set environment variables
export CONTAINER_EP=unix:///var/run/runtime-launcher.sock
mkdir -p /tmp/yr_sessions
mkdir -p /var/log/supervisor
mkdir -p /var/run

# Start supervisord
exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf -n
