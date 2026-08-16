#!/usr/bin/env bash
# Provisions the Address Book dashboard into a running OpenObserve instance.
#
# Usage:
#   OO_BASE=http://localhost:5080 OO_EMAIL=admin@example.com OO_PASSWORD='Admin@12345' \
#     ./install-dashboard.sh
set -euo pipefail

OO_BASE="${OO_BASE:-http://localhost:5080}"
OO_EMAIL="${OO_EMAIL:-admin@example.com}"
OO_PASSWORD="${OO_PASSWORD:-Admin@12345}"
OO_ORG="${OO_ORG:-default}"
DASHBOARD_JSON="${DASHBOARD_JSON:-$(dirname "$0")/addressbook.json}"

auth=$(printf '%s:%s' "$OO_EMAIL" "$OO_PASSWORD" | base64)

# The create endpoint requires the v8 schema. `dashboardId` and
# `defaultDatetimeDuration` are not part of the create body, so strip them.
tmp=$(mktemp)
python3 - "$DASHBOARD_JSON" "$tmp" <<'PY'
import json, sys
src, dst = sys.argv[1], sys.argv[2]
d = json.load(open(src))
for k in ("dashboardId", "defaultDatetimeDuration", "created"):
    d.pop(k, None)
json.dump(d, open(dst, "w"))
PY

echo "Provisioning dashboard into $OO_BASE ..."
resp=$(curl -sf -X POST \
  -H "Authorization: Basic $auth" \
  -H "Content-Type: application/json; charset=UTF-8" \
  --data-binary "@$tmp" \
  "$OO_BASE/api/$OO_ORG/dashboards?folder=default")

rm -f "$tmp"

# Extract the created dashboard id
did=$(printf '%s' "$resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['v8']['dashboardId'])")
echo "Provisioned dashboard id: $did"
echo "Open it at $OO_BASE/web/dashboards/$did"
