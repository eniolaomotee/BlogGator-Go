#!/bin/bash
set -euo pipefall

cd sql/schema

goose postgres "$DATABASE_URL" up