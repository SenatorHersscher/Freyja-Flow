#!/usr/bin/env bash
set -e

echo "🔮 Freyja Flow - Derleme Başlatılıyor..."

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
BIN_DIR="$ROOT_DIR/bin"

mkdir -p "$BIN_DIR"
mkdir -p "$BACKEND_DIR/frontend_embed"

echo "📦 Frontend dosyaları senkronize ediliyor..."
rm -rf "$BACKEND_DIR/frontend_embed"
cp -r "$ROOT_DIR/frontend" "$BACKEND_DIR/frontend_embed"

cd "$BACKEND_DIR"

echo "🐧 Linux (x86_64) derleniyor..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$BIN_DIR/freyja-flow-linux-x64" main.go

echo "🪟 Windows (x86_64 .exe) derleniyor..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$BIN_DIR/freyja-flow-windows-x64.exe" main.go

echo "✅ Derleme tamamlandı! Çıktılar:"
ls -lh "$BIN_DIR"
