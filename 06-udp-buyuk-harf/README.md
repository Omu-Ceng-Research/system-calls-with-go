# 06 — UDP soketi: bağlantısız sunucu ve beş istemci

Her istemci saniyede bir kez kendi harfini (aaaaaaaaaa, bbbbbbbbbb …) göndermekte, sunucu
büyük harfe çevirip geri yollamaktadır. 2023-24 dönemindeki Python örneği ve Go karşılığı
aynı klasörde yer almaktadır. İkisi de **12000** (UDP) portunu kullanmaktadır.

| Dosya | Açıklama |
|---|---|
| [`main.go`](main.go) | Go sürümü (sunucu, tek istemci ve beş istemci tek programda) |
| [`python/server.py`](python/server.py) | 2023-24 Python sunucusu |
| [`python/client.py`](python/client.py) … [`client5.py`](python/client5.py) | 2023-24 Python istemcileri (a … e harfleri) |

Bütün dosyalarda soket satırlarının yanında, o satırın yaptığı sistem çağrıları `→` ile
yazılmıştır.

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
| `ders 06 sunucu` | `go run ./06-udp-buyuk-harf sunucu` |
| `python3 /kod/06-udp-buyuk-harf/python/...` | `python3 06-udp-buyuk-harf/python/...` |
| `cd /calisma` ve `strace ... ./d06 ...` | `go build -o d06 ./06-udp-buyuk-harf` ve `strace ... ./d06 ...` |

Linux'ta strace ve nc paketleri kurulu olmalıdır. Windows ve macOS'ta bazı sistem
çağrısı deneyleri (strace) yalnızca Docker'da yapılabilmektedir.

## Go

```bash
# 1. terminal: sunucu
ders 06 sunucu

# 2. terminal: tek istemci, a harfi (Ctrl+C ile durdurulana kadar)
ders 06 istemci a

# 2. terminal: tek istemci, b harfi, 5 mesaj gönderip çıkar
ders 06 istemci b -adet 5

# 2. terminal: beş istemci (a … e) aynı anda, tek programda
ders 06 coklu -adet 3
```

## Python (2023-24)

Python istemcileri durdurulana kadar çalışmaktadır. Her biri için ayrı bir terminal
açabilir ya da hepsini arka planda başlatabilirsiniz.

```bash
# 1. terminal: sunucu
cd /kod/06-udp-buyuk-harf/python
python3 -u server.py

# 2. terminal: tek istemci
cd /kod/06-udp-buyuk-harf/python
python3 -u client.py

# 2. terminal: beş istemci arka planda, 5 saniye sonra hepsi durdurulur
cd /kod/06-udp-buyuk-harf/python
for c in client.py client2.py client3.py client4.py client5.py; do timeout 5 python3 -u $c & done; wait
```

`-u` bayrağı Python'un çıktıyı tamponda bekletmemesini, her satırı hemen basmasını
sağlamaktadır.

## Karışık: Python ↔ Go

```bash
# Python sunucu + Go istemciler
python3 -u /kod/06-udp-buyuk-harf/python/server.py        # 1. terminal
ders 06 coklu -adet 3                                      # 2. terminal

# Go sunucu + Python istemci
ders 06 sunucu                                             # 1. terminal
python3 -u /kod/06-udp-buyuk-harf/python/client.py         # 2. terminal
```

## Denenmesi gerekenler

```bash
# Sunucu kapalıyken istemci: Python sonsuza kadar bekler, Go 2 saniyede "cevap yok" der
python3 -u /kod/06-udp-buyuk-harf/python/client.py
ders 06 istemci a -adet 3

# İstemciler çalışırken sunucuyu Ctrl+C ile durdurup yeniden başlatınız:
# UDP'de bağlantı olmadığı için istemciler kaldıkları yerden devam etmektedir.
```

## Sistem çağrılarını gözlemlemek

```bash
cd /calisma
strace -f -e trace=network,read,write,epoll_pwait ./d06 sunucu
strace -f -e trace=network,read,write,epoll_pwait ./d06 istemci a -adet 2

cd /kod/06-udp-buyuk-harf/python
strace -f -e trace=network,read,write python3 -u server.py
strace -f -e trace=network,read,write,clock_nanosleep python3 -u client.py
```

Sunucunun strace çıktısında `listen` ve `accept` bulunmadığını; her paketin
`recvfrom` ile, gönderenin adresiyle birlikte alındığını doğrulayınız.
