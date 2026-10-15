#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
SERVICE_NAME="job-radar.service"
USER_SYSTEMD_DIR="${HOME}/.config/systemd/user"

echo "==> Compilando binário otimizado do Job Radar..."
cd "${PROJECT_DIR}"
go build -ldflags="-s -w" -o job-radar ./cmd/job-radar

echo "==> Configurando diretório systemd do usuário..."
mkdir -p "${USER_SYSTEMD_DIR}"

echo "==> Copiando arquivo de serviço..."
cp "${PROJECT_DIR}/scripts/${SERVICE_NAME}" "${USER_SYSTEMD_DIR}/${SERVICE_NAME}"

echo "==> Recarregando systemd --user..."
systemctl --user daemon-reload || true

echo "==> Habilitando serviço para iniciar automaticamente com seu usuário..."
systemctl --user enable "${SERVICE_NAME}" || true

echo ""
echo "✅ Instalação concluída com sucesso!"
echo "Comandos úteis para gerenciar seu daemon:"
echo "  Iniciar agora : systemctl --user start ${SERVICE_NAME}"
echo "  Ver status    : systemctl --user status ${SERVICE_NAME}"
echo "  Ver logs      : journalctl --user -u ${SERVICE_NAME} -f"
echo "  Parar serviço : systemctl --user stop ${SERVICE_NAME}"
