#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LIB_DIR="${ROOT_DIR}/../ebook-mechanic-lib"
if [[ -n "${OUT_DIR:-}" ]]; then
  OUT_DIR="${OUT_DIR}"
elif [[ -n "${FIXTURES_REPAIR:-}" ]]; then
  OUT_DIR="${FIXTURES_REPAIR}"
elif [[ -n "${1:-}" ]]; then
  OUT_DIR="${1}"
else
  OUT_DIR="${ROOT_DIR}/test-library/repair-fixtures"
fi

epub_src_dir="${LIB_DIR}/testdata/epub/invalid"
pdf_src_dir="${LIB_DIR}/testdata/pdf/invalid"

mkdir -p "${OUT_DIR}/epub" "${OUT_DIR}/pdf"

epub_fixtures=(
  "epub_mimetype_content__fix-mimetype-content.epub|${epub_src_dir}/wrong_mimetype.epub"
  "epub_mimetype_compressed__fix-mimetype-uncompressed.epub|${epub_src_dir}/mimetype_compressed.epub"
  "epub_mimetype_not_first__fix-mimetype-order.epub|${epub_src_dir}/mimetype_not_first.epub"
  "epub_container_missing__create-container-xml.epub|${epub_src_dir}/no_container.epub"
  "epub_content_missing_doctype__add-doctype.epub|${epub_src_dir}/invalid_content_document.epub"
  "epub_opf_missing_title__add-title.epub|${epub_src_dir}/missing_title.epub"
  "epub_opf_missing_identifier__add-identifier.epub|${epub_src_dir}/missing_identifier.epub"
  "epub_opf_missing_language__add-language.epub|${epub_src_dir}/missing_language.epub"
  "epub_opf_missing_modified__add-modified.epub|${epub_src_dir}/missing_modified.epub"
  "epub_opf_missing_nav__add-nav-document.epub|${epub_src_dir}/missing_nav_document.epub"
)

pdf_fixtures=(
  "pdf_missing_eof__append-eof.pdf|${pdf_src_dir}/no_eof.pdf"
  "pdf_missing_startxref__recompute-startxref.pdf|${pdf_src_dir}/no_startxref.pdf"
)

copy_fixture() {
  local dest_name="$1"
  local src_path="$2"
  local dest_path="$3"

  if [[ ! -f "${src_path}" ]]; then
    echo "Missing source fixture: ${src_path}" >&2
    exit 1
  fi

  cp -f "${src_path}" "${dest_path}/${dest_name}"
}

for entry in "${epub_fixtures[@]}"; do
  IFS="|" read -r dest_name src_path <<<"${entry}"
  copy_fixture "${dest_name}" "${src_path}" "${OUT_DIR}/epub"
done

for entry in "${pdf_fixtures[@]}"; do
  IFS="|" read -r dest_name src_path <<<"${entry}"
  copy_fixture "${dest_name}" "${src_path}" "${OUT_DIR}/pdf"
done

echo "Repair fixtures created in ${OUT_DIR}"
