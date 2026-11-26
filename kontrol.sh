#!/usr/bin/env bash
# Bütün örnekleri derler ve her birini kısa, etkileşimsiz bir biçimde çalıştırır.
# Sunucu/istemci örneklerinde sunucu arka planda başlatılıp istemci bitince kapatılır.
# Gerçek Gmail'e bağlanan adımlar çalıştırılmaz; SMTP örnekleri sahte sunucuyla denenir.
#
# Kullanım (Docker ortamının içinde ya da Linux'ta): ./kontrol.sh
set -euo pipefail
cd "$(dirname "$0")"
KOK=$PWD
CALISMA=$(mktemp -d)
trap 'kill $(jobs -p) 2>/dev/null || true; rm -rf "$CALISMA"' EXIT

echo "== $(go version)"
fark=$(gofmt -l .)
if [[ -n "$fark" ]]; then
    echo "Biçimlendirilmemiş dosyalar (gofmt -w . ile düzeltiniz):" >&2
    echo "$fark" >&2
    exit 1
fi
go vet ./...

for klasor in [0-9][0-9]-*/; do
    no=${klasor:0:2}
    go build -o "$CALISMA/d$no" "./$klasor"
done
cd "$CALISMA"

calistir() { # calistir <başlık> <komut...>
    echo
    echo "################ $1"
    shift
    "$@"
}
sunucu() { # sunucu <komut...>: arka planda başlatır, dinlemeye başlamasını bekler
    "$@" > sunucu.log 2>&1 &
    sleep 0.5
}
durdur() {
    kill %1 2>/dev/null || true
    wait 2>/dev/null || true
    cat sunucu.log
}

# Python sunucusu SO_REUSEADDR açmadığı için Go örneklerinden ÖNCE denenmektedir
# (Go sunucusundan kalan TIME_WAIT bağlantıları portu ~60 sn meşgul tutmaktadır).
calistir "05 Python" true
sunucu python3 "$KOK/05-tcp-buyuk-harf/python/tcp-server.py"
echo "istanbul" | ./d05 istemci | tail -1; durdur > /dev/null
sunucu ./d05 sunucu
echo "istanbul" | python3 "$KOK/05-tcp-buyuk-harf/python/tcp-client.py" | tail -1; durdur > /dev/null

calistir "06 Python" true
sunucu python3 -u "$KOK/06-udp-buyuk-harf/python/server.py"
./d06 coklu -adet 1; durdur > /dev/null
sunucu ./d06 sunucu
timeout 1.5 python3 -u "$KOK/06-udp-buyuk-harf/python/client.py" || true; durdur > /dev/null

calistir 05 true; sunucu ./d05 sunucu
echo "istanbul ığüşöç" | ./d05 istemci; durdur

calistir 06 true; sunucu ./d06 sunucu
./d06 coklu -adet 1; durdur

calistir 07 true; sunucu ./d07 sunucu
for y in tek parcali kilitli; do ./d07 istemci -yazim $y; sleep 0.3; done; durdur

calistir 08 true; sunucu ./d08 sunucu
for b in gob json ikili; do ./d08 istemci -bicim $b; sleep 0.2; done; durdur

calistir "09 + 10" true; sunucu ./d09 sahte-sunucu
./d09 gonder -sunucu 127.0.0.1:2525 -tls=false -alici a@ornek.com
./d10 renkli -sunucu 127.0.0.1:2525 -alici a@ornek.com
./d10 ekli -sunucu 127.0.0.1:2525 -alici a@ornek.com -dosya "$KOK/go.mod"
sleep 0.3; durdur > /dev/null

calistir 11 ./d11
calistir 12 ./d12
echo
echo "TAMAM: bütün örnekler çalıştı."
