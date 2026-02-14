#!/bin/sh
set -e

export CONFIG_PATH="${CONFIG_PATH:-/app/resources/config/}"

exec /app/sweetshop "$@"
