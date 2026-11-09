#!/usr/bin/env bash
#
# syncFTP release paketleri — Windows'tan çalışır (CGO yok, imza rcodesign ile).
# Çıktı dist/: platform başına bir zip + Windows exe.
#
# macOS imza kademeleri:
#   0  rcodesign yok            → betik durur (imzasız paket "hasarlı" görünür)
#   1  MAC_SIGN=1               → ad-hoc imza, noterlik yok (hızlı yerel deneme)
#   2  Developer ID + noterlik  → çift tıkla açılır, uyarı yok (varsayılan)
# Anahtarlar REPO DIŞINDA: APPLE_SIGN_DIR (varsayılan ~/.agentic-signing/apple)
#   devid.key, devid.cer, api-key.json
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION="$(sed -n 's/^const appVersion = "\(.*\)"/\1/p' cmd/syncftp/cmd_gui_api.go | tr -d '\r')"
BUNDLE_ID="com.mustafaomereser.syncftp"
OUT="${OUT:-dist}"
APPLE_SIGN_DIR="${APPLE_SIGN_DIR:-$HOME/.agentic-signing/apple}"
NOTARY_WAIT="${NOTARY_WAIT:-7200}"
MAC_SIGN="${MAC_SIGN:-2}"
HELPER_DIR="${HELPER_DIR:-packaging/helper}"
LD="-s -w"

command -v rcodesign >/dev/null || { echo "HATA: rcodesign yok" >&2; exit 1; }
if [ "$MAC_SIGN" = 2 ]; then
  for f in devid.key devid.cer api-key.json; do
    [ -f "$APPLE_SIGN_DIR/$f" ] || { echo "HATA: $APPLE_SIGN_DIR/$f yok" >&2; exit 1; }
  done
fi

PY=""
for c in python python3 py; do
  if command -v "$c" >/dev/null && "$c" -c 'import zipfile' >/dev/null 2>&1; then PY="$c"; break; fi
done
[ -n "$PY" ] || { echo "HATA: python yok" >&2; exit 1; }

echo "syncFTP $VERSION"
if [ "${SKIP_TESTS:-0}" != 1 ]; then
  echo "  vet + test"
  go vet ./...
  go test ./...
fi

rm -rf "$OUT"; mkdir -p "$OUT/stage"

build() { # goos goarch çıktı
  echo "  derleme: $1/$2"
  CGO_ENABLED=0 GOOS="$1" GOARCH="$2" go build -trimpath -ldflags="$LD" -o "$3" ./cmd/syncftp/
}

build windows amd64 "$OUT/syncFTP-$VERSION-windows-amd64.exe"

# ── macOS ────────────────────────────────────────────────────────────────────
NOTARIZE=()
make_app() { # goarch klasör
  local root="$OUT/stage/$2" app="$OUT/stage/$2/syncFTP.app"
  mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
  build darwin "$1" "$app/Contents/MacOS/syncftp"
  local exec_name=syncftp helper="$HELPER_DIR/syncftp-window-darwin-$1"
  if [ -f "$helper" ]; then
    cp "$helper" "$app/Contents/MacOS/syncftp-window"; chmod +x "$app/Contents/MacOS/syncftp-window"; exec_name=syncftp-window
  elif [ "${REQUIRE_HELPERS:-0}" = 1 ]; then echo "HATA: $helper yok" >&2; exit 1; fi
  cp packaging/syncftp.icns "$app/Contents/Resources/syncftp.icns"
  cat > "$app/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key><string>syncFTP</string>
  <key>CFBundleDisplayName</key><string>syncFTP</string>
  <key>CFBundleIdentifier</key><string>$BUNDLE_ID</string>
  <key>CFBundleVersion</key><string>$VERSION</string>
  <key>CFBundleShortVersionString</key><string>$VERSION</string>
  <key>CFBundleExecutable</key><string>$exec_name</string>
  <key>CFBundleIconFile</key><string>syncftp</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>LSUIElement</key><false/>
  <key>NSHighResolutionCapable</key><true/>
  <key>NSAppTransportSecurity</key><dict><key>NSAllowsLocalNetworking</key><true/></dict>
</dict>
</plist>
PLIST
  if [ "$MAC_SIGN" = 2 ]; then
    rcodesign sign --pem-file "$APPLE_SIGN_DIR/devid.key" --certificate-der-file "$APPLE_SIGN_DIR/devid.cer" \
      --code-signature-flags runtime "$app" >/dev/null
    NOTARIZE+=("$app")
  else
    rcodesign sign "$app" >/dev/null
  fi
  [ -s "$app/Contents/_CodeSignature/CodeResources" ] || { echo "HATA: imza oluşmadı: $app" >&2; exit 1; }
  echo "  imzalandı: $2"
}
make_app arm64 macos-apple-silicon
make_app amd64 macos-intel

if [ "$MAC_SIGN" = 2 ]; then
  key="$APPLE_SIGN_DIR/api-key.json"; ids=()
  for app in "${NOTARIZE[@]}"; do
    out="$(rcodesign notary-submit --api-key-file "$key" "$app" 2>&1)" || { echo "$out" | tail -20 >&2; exit 1; }
    id="$(echo "$out" | grep -oiE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' | tail -1)"
    [ -n "$id" ] || { echo "$out" | tail -20 >&2; echo "HATA: gönderim kimliği yok" >&2; exit 1; }
    echo "  noterliğe gönderildi: $app ($id)"
    ids+=("$id")
  done
  for i in "${!NOTARIZE[@]}"; do
    app="${NOTARIZE[$i]}"; id="${ids[$i]}"
    echo "  bekleniyor: $id"
    out="$(rcodesign notary-wait --api-key-file "$key" --max-wait-seconds "$NOTARY_WAIT" "$id" 2>&1)" || {
      echo "$out" | tail -8 >&2; echo "HATA: noterlik tamamlanmadı ($id)" >&2; exit 1; }
    rcodesign staple "$app" >/dev/null
    echo "  onaylandı: $app"
  done
fi

# ── Linux ────────────────────────────────────────────────────────────────────
make_linux() { # goarch
  local d="$OUT/stage/linux-$1"
  mkdir -p "$d"
  build linux "$1" "$d/syncftp"
  if [ -f "$HELPER_DIR/syncftp-window-linux-$1" ]; then
    cp "$HELPER_DIR/syncftp-window-linux-$1" "$d/syncftp-window"
  elif [ "${REQUIRE_HELPERS:-0}" = 1 ]; then echo "HATA: Linux pencere programı yok ($1)" >&2; exit 1; fi
  cp packaging/syncftp.png "$d/syncftp.png"
  cat > "$d/syncftp.desktop" <<'DESKTOP'
[Desktop Entry]
Type=Application
Name=syncFTP
Comment=Deploy changed files to FTP / FTPS / SFTP servers
Exec=syncftp gui
Icon=syncftp
Terminal=false
Categories=Development;Network;
DESKTOP
  cat > "$d/install.sh" <<'INSTALL'
#!/usr/bin/env bash
# Kullanıcı hesabına kurar (yönetici yetkisi gerekmez): komut + uygulama menüsü.
set -euo pipefail
src="$(cd "$(dirname "$0")" && pwd)"
mkdir -p "$HOME/.local/bin" "$HOME/.local/share/applications" "$HOME/.local/share/icons/hicolor/512x512/apps"
install -m 755 "$src/syncftp" "$HOME/.local/bin/syncftp"
[ -f "$src/syncftp-window" ] && install -m 755 "$src/syncftp-window" "$HOME/.local/bin/syncftp-window"
install -m 644 "$src/syncftp.png" "$HOME/.local/share/icons/hicolor/512x512/apps/syncftp.png"
sed "s|^Exec=.*|Exec=$HOME/.local/bin/syncftp gui|" "$src/syncftp.desktop" > "$HOME/.local/share/applications/syncftp.desktop"
echo "Kuruldu: uygulama menüsünde syncFTP, terminalde ~/.local/bin/syncftp"
echo "Uygulama penceresi için libwebkit2gtk-4.1 gerekir; yoksa panel Chrome ile açılır."
INSTALL
}
make_linux amd64
make_linux arm64

# ── Zip'ler: Unix izinleriyle (Windows'ta yazılan zip varsayılan MS-DOS olur, program çalıştırılamaz çıkar) ──
"$PY" -I - "$OUT" "$VERSION" <<'PYZIP'
import os, sys, zipfile
out, ver = sys.argv[1], sys.argv[2]
stage = os.path.join(out, 'stage')
def is_exe(rel):
    base = rel.split('/')[-1]
    return '/Contents/MacOS/' in rel or base in ('syncftp', 'syncftp-window', 'install.sh')
for name in sorted(os.listdir(stage)):
    src = os.path.join(stage, name)
    dest = os.path.join(out, f'syncFTP-{ver}-{name}.zip')
    with zipfile.ZipFile(dest, 'w', zipfile.ZIP_DEFLATED) as z:
        for root, dirs, files in os.walk(src):
            dirs.sort()
            rel_root = os.path.relpath(root, src).replace(os.sep, '/')
            if rel_root != '.':
                d = zipfile.ZipInfo(rel_root + '/'); d.create_system = 3; d.external_attr = (0o40755 << 16) | 0x10
                z.writestr(d, b'')
            for f in sorted(files):
                p = os.path.join(root, f)
                rel = os.path.relpath(p, src).replace(os.sep, '/')
                i = zipfile.ZipInfo.from_file(p, rel); i.create_system = 3; i.compress_type = zipfile.ZIP_DEFLATED
                i.external_attr = (0o100755 if is_exe(rel) else 0o100644) << 16
                with open(p, 'rb') as fh:
                    z.writestr(i, fh.read())
    with zipfile.ZipFile(dest) as z:
        assert z.testzip() is None, dest
        bad = [n.filename for n in z.infolist() if is_exe(n.filename) and not (n.external_attr >> 16) & 0o111]
        assert not bad, bad
    print('  zip:', os.path.basename(dest))
PYZIP
rm -rf "$OUT/stage"

# Güncelleme imzası: uygulama yalnızca bu imzayı tutan paketi kurar (cmd/syncftp/update.go)
KEYFILE="$HOME/.agentic-signing/syncftp-update.key"
if [ -z "${UPDATE_SIGNING_KEY:-}" ] && [ -f "$KEYFILE" ]; then UPDATE_SIGNING_KEY="$(cat "$KEYFILE")"; fi
[ -n "${UPDATE_SIGNING_KEY:-}" ] || { echo "HATA: güncelleme imza anahtarı yok" >&2; exit 1; }
( cd "$OUT" && sha256sum syncFTP-* > checksums.txt )
UPDATE_SIGNING_KEY="$UPDATE_SIGNING_KEY" go run ./cmd/signrelease "$OUT/checksums.txt"
echo "  imzalandı: checksums.txt"
echo; ls -lh "$OUT" | awk 'NR>1 {printf "  %-46s %s\n", $9, $5}'
