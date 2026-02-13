#!/bin/bash
set -euo pipefail

cd sql/schema

goose postgres "$DATABASE_URL" up