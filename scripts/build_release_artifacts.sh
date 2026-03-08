#!/usr/bin/env bash

set -euo pipefail

VERSION="${1:-v1.0.0}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${REPO_ROOT}/bin/releases/${VERSION}"
PACKAGE="./cmd/swapchess"
VERSION_PKG="github.com/divijg19/Swapchess/internal/app"

TARGETS=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
  "windows amd64"
)

rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

ARCHIVES=()

for target in "${TARGETS[@]}"; do
  read -r GOOS GOARCH <<<"${target}"
  TARGET_DIR="${DIST_DIR}/${GOOS}-${GOARCH}"
  mkdir -p "${TARGET_DIR}"

  OUTPUT_NAME="swapchess"
  if [ "${GOOS}" = "windows" ]; then
    OUTPUT_NAME="swapchess.exe"
  fi

  echo "Building ${GOOS}/${GOARCH}"
  (
    cd "${REPO_ROOT}"
    GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=0 \
      go build -trimpath \
      -ldflags "-s -w -X ${VERSION_PKG}.Version=${VERSION}" \
      -o "${TARGET_DIR}/${OUTPUT_NAME}" \
      "${PACKAGE}"
  )

  if [ "${GOOS}" = "windows" ]; then
    ARCHIVE_PATH="${DIST_DIR}/swapchess_${VERSION}_${GOOS}_${GOARCH}.zip"
    (
      cd "${TARGET_DIR}"
      zip -q -r "${ARCHIVE_PATH}" .
    )
  else
    ARCHIVE_PATH="${DIST_DIR}/swapchess_${VERSION}_${GOOS}_${GOARCH}.tar.gz"
    tar -C "${TARGET_DIR}" -czf "${ARCHIVE_PATH}" .
  fi
  ARCHIVES+=("${ARCHIVE_PATH}")
done

(
  cd "${DIST_DIR}"
  sha256sum "${ARCHIVES[@]}" | sed "s#${DIST_DIR}/##" > SHA256SUMS.txt
)

echo "Release artifacts written to ${DIST_DIR}"