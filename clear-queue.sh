#!/usr/bin/env bash
set -euo pipefail

db="${XDG_DATA_HOME:-$HOME/.local/share}/bit-pro/bit.db"

n="$(sqlite3 "$db" 'DELETE FROM queue; SELECT changes();')"
echo "cleared $n queue rows"
