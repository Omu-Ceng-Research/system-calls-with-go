# 09 — SMTP ile e-posta: kütüphanesiz, doğrudan soket üzerinden

2023-24 dönemindeki "telnet ile e-posta gönderme" uygulaması ve "kütüphane kullanmadan
e-posta istemcisi yazınız" alıştırmasıdır. Go programı SMTP komutlarını tek tek sokete
yazmakta ve sunucunun cevaplarını ekrana basmaktadır. Gmail hesabı olmadan denemek için
programın içinde küçük bir **sahte SMTP sunucusu** (port **2525**) bulunmaktadır.

| Dosya | Açıklama |
|---|---|
| [`main.go`](main.go) | Go istemcisi (`gonder`) ve sahte sunucu (`sahte-sunucu`) |
| [`telnet/smtp-adimlari.txt`](telnet/smtp-adimlari.txt) | 2023-24: openssl ile elle yapılan SMTP oturumunun adımları |
| [`telnet/gmail-uygulama-sifresi.md`](telnet/gmail-uygulama-sifresi.md) | 2023-24: Gmail uygulama şifresi alma adımları |

Bütün dosyalarda soket satırlarının (ve elle yazılan adımların) yanında, yapılan sistem
çağrıları `→` ile yazılmıştır.

## Ortamı açmak

Komutlar Docker ortamının içinde çalıştırılmaktadır. Repo klasöründe:

```bash
./baslat.sh          # Windows: baslat.cmd dosyasına çift tıklayınız
```

**İkinci terminal** için aynı komutu başka bir terminalde bir kez daha çalıştırınız; açık
olan ortama ikinci bir kabuk açılmaktadır. Sunucuyu birinci, istemciyi ikinci terminalde
çalıştırınız. Bir programı durdurmak için `Ctrl+C` kullanınız.

Go bilgisayarınızda kuruluysa Docker olmadan da çalıştırabilirsiniz: aşağıdaki
`ders NN ...` komutları yerine repo klasöründe `go run ./KLASÖR ...` yazınız.

## Go, sahte sunucuyla (hesap gerekmez)

```bash
# 1. terminal: sahte sunucu (gelen e-postayı ekrana basar, kimseye göndermez)
ders 09 sahte-sunucu

# 2. terminal: istemci
ders 09 gonder -sunucu 127.0.0.1:2525 -tls=false -alici ogrenci@ornek.com

# konu ve gövde ile
ders 09 gonder -sunucu 127.0.0.1:2525 -tls=false -alici ogrenci@ornek.com \
    -konu "Deneme" -mesaj "Merhaba, bu bir denemedir."
```

## Elle, telnet gibi (2023-24 yöntemi)

Sahte sunucuya `nc` ile bağlanıp komutları kendiniz yazınız (`-C`: satır sonu `\r\n`):

```bash
ders 09 sahte-sunucu          # 1. terminal
nc -C 127.0.0.1 2525          # 2. terminal, ardından sırayla:
```

```
EHLO localhost
MAIL FROM:<ben@ornek.com>
RCPT TO:<sen@ornek.com>
DATA
Subject: Deneme
Merhaba.
.
QUIT
```

Gerçek Gmail ile, 2023-24'teki gibi (adımlar [`telnet/smtp-adimlari.txt`](telnet/smtp-adimlari.txt) dosyasındadır):

```bash
openssl s_client -connect smtp.gmail.com:465 -crlf -ign_eof
```

## Go, gerçek Gmail ile

Önce bir uygulama şifresi alınız ([`telnet/gmail-uygulama-sifresi.md`](telnet/gmail-uygulama-sifresi.md)).
Normal Gmail şifresi kabul edilmemektedir. Şifreyi koda yazmayınız; ortam değişkeniyle
veriniz:

```bash
export GMAIL_ADRES=adiniz@gmail.com
export GMAIL_SIFRE='abcd efgh ijkl mnop'      # uygulama şifresi
ders 09 gonder -alici arkadasiniz@ornek.com
```

## Sistem çağrılarını gözlemlemek

```bash
cd /calisma
strace -f -e trace=network,read,write,close ./d09 gonder -sunucu 127.0.0.1:2525 -tls=false -alici a@b.c
strace -f -e trace=network,read,write openssl s_client -connect smtp.gmail.com:465 -crlf -ign_eof
```

Birinci komutta her SMTP komutu düz metin olarak bir `write` satırında görünmektedir.
İkincisinde ise sokete giden veri şifreli olduğu için okunamamaktadır; yazdığınız komut
yalnızca klavyeden okunan `read(0, ...)` satırında görünmektedir.
