# 10 — SMTP kütüphanesiyle renkli ve ek dosyalı e-posta

2023-24 dönemindeki 4. hafta alıştırmasının Go karşılığıdır: renkli kelimeler içeren ve
ekinde bir metin dosyası bulunan e-postalar, Go'nun `net/smtp` paketiyle
gönderilmektedir. 2023-24'te bu alıştırma için hazır kod verilmemişti; bu yüzden klasörde
yalnızca Go sürümü bulunmaktadır.

| Dosya | Açıklama |
|---|---|
| [`main.go`](main.go) | `renkli` ve `ekli` e-posta gönderen Go programı |

`SendMail` satırının yanında, kütüphanenin arka planda yaptığı sistem çağrıları `→`
ile yazılmıştır.

## Ortamı açmak

Komutlar Docker ortamının içinde çalıştırılmaktadır. Repo klasöründe:

```bash
./baslat.sh          # Windows: baslat.cmd dosyasına çift tıklayınız
```

**İkinci terminal** için aynı komutu başka bir terminalde bir kez daha çalıştırınız; açık
olan ortama ikinci bir kabuk açılmaktadır. Sunucuyu birinci, istemciyi ikinci terminalde
çalıştırınız. Bir programı durdurmak için `Ctrl+C` kullanınız.

### Docker olmadan (Go bilgisayarınızda kuruluysa)

Komutlar **repo klasöründe** şu karşılıklarıyla çalıştırılmaktadır:

| Docker'da | Bilgisayarınızda |
|---|---|
| `ders 10 ...` | `go run ./10-smtp-kutuphane ...` |
| `-dosya /kod/go.mod` | `-dosya go.mod` |
| `cd /calisma` ve `strace ... ./d10 ...` | `go build -o d10 ./10-smtp-kutuphane` ve `strace ... ./d10 ...` |

Linux'ta strace ve nc paketleri kurulu olmalıdır. Windows ve macOS'ta bazı sistem
çağrısı deneyleri (strace) yalnızca Docker'da yapılabilmektedir.

## Go, sahte sunucuyla (hesap gerekmez)

Sahte sunucu örnek 09'dadır:

```bash
# 1. terminal
ders 09 sahte-sunucu

# 2. terminal
ders 10 renkli -sunucu 127.0.0.1:2525 -alici ogrenci@ornek.com
ders 10 ekli   -sunucu 127.0.0.1:2525 -alici ogrenci@ornek.com -dosya /kod/go.mod
```

Sahte sunucunun ekranında e-postanın ham hâlini (HTML gövde, MIME sınırları ve base64 ile
kodlanmış ek) görebilirsiniz.

## Go, gerçek Gmail ile

```bash
export GMAIL_ADRES=adiniz@gmail.com
export GMAIL_SIFRE='abcd efgh ijkl mnop'      # uygulama şifresi (bkz. örnek 09)
ders 10 renkli -alici arkadasiniz@ornek.com
ders 10 ekli   -alici arkadasiniz@ornek.com -dosya /kod/README.md
```

## Sistem çağrılarını gözlemlemek

```bash
cd /calisma
strace -f -e trace=network,read,write,close ./d10 renkli -sunucu 127.0.0.1:2525 -alici a@b.c
```

Çıktıyı örnek 09'unkiyle karşılaştırınız: kütüphane kullanmak sistem çağrılarını
değiştirmemektedir; aynı `connect`, `write` ve `read` çağrıları yapılmaktadır.
