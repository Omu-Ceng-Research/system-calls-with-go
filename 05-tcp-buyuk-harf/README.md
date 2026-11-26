# 05 — TCP soketi: büyük harfe çeviren sunucu

İstemci bir cümle göndermekte, sunucu cümleyi büyük harfe çevirip geri yollamaktadır.
2023-24 dönemindeki Python örneği ve Go karşılığı aynı klasörde yer almaktadır. İkisi de
**12001** portunu kullandığı için birbirleriyle konuşabilmektedir.

| Dosya | Açıklama |
|---|---|
| [`main.go`](main.go) | Go sürümü (sunucu ve istemci tek programda) |
| [`python/tcp-server.py`](python/tcp-server.py) | 2023-24 Python sunucusu |
| [`python/tcp-client.py`](python/tcp-client.py) | 2023-24 Python istemcisi |

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
| `ders 05 sunucu` | `go run ./05-tcp-buyuk-harf sunucu` |
| `python3 /kod/05-tcp-buyuk-harf/python/...` | `python3 05-tcp-buyuk-harf/python/...` |
| `cd /calisma` ve `strace ... ./d05 ...` | `go build -o d05 ./05-tcp-buyuk-harf` ve `strace ... ./d05 ...` |

Linux'ta strace ve nc paketleri kurulu olmalıdır. Windows ve macOS'ta bazı sistem
çağrısı deneyleri (strace) yalnızca Docker'da yapılabilmektedir.

## Go

```bash
# 1. terminal: sunucu
ders 05 sunucu

# 2. terminal: istemci (cümleyi klavyeden yazınız)
ders 05 istemci

# 2. terminal: istemci, cümle komut satırından
echo "istanbul ığüşöç" | ders 05 istemci

# 2. terminal: başka bir makinedeki sunucuya
ders 05 istemci 10.0.0.12
```

## Python (2023-24)

```bash
# 1. terminal: sunucu
cd /kod/05-tcp-buyuk-harf/python
python3 tcp-server.py

# 2. terminal: istemci
cd /kod/05-tcp-buyuk-harf/python
python3 tcp-client.py
```

## Karışık: Python ↔ Go

Protokol aynıdır (ham baytlar, tek mesaj). Bir tarafı Python, diğerini Go ile çalıştırınız:

```bash
# Python sunucu + Go istemci
python3 /kod/05-tcp-buyuk-harf/python/tcp-server.py      # 1. terminal
echo "istanbul" | ders 05 istemci                         # 2. terminal → ISTANBUL

# Go sunucu + Python istemci
ders 05 sunucu                                            # 1. terminal
echo "istanbul" | python3 /kod/05-tcp-buyuk-harf/python/tcp-client.py   # → İSTANBUL
```

Sonuçtaki farka dikkat ediniz: Python'un `upper()` fonksiyonu Türkçe kurallarını
bilmemektedir ("ISTANBUL"), Go sürümü ise `unicode.TurkishCase` kullanmaktadır ("İSTANBUL").

İki sunucu aynı anda çalıştırılamaz: ikisi de 12001 portunu dinlediği için ikinci sunucu
`address already in use` hatası vermektedir.

Python sunucusu, bir sunucu kapandıktan sonraki ~60 saniye içinde başlatılırsa
`OSError: [Errno 98] Address already in use` hatası verebilmektedir. Bunun nedeni
kapanan bağlantıların TIME_WAIT durumunda beklemesi ve Python'un `SO_REUSEADDR`
seçeneğini açmamasıdır. Go sunucusu bu seçeneği kendiliğinden açtığı için hemen yeniden
başlatılabilmektedir. Ayrıntı ve deney için: [12-so-reuseaddr](../12-so-reuseaddr).
Bir dakika bekleyiniz ya da `tcp-server.py` içindeki açıklamada gösterilen
`setsockopt` satırını ekleyiniz.

## Sistem çağrılarını gözlemlemek

```bash
cd /calisma
strace -f -e trace=network,read,write,close,epoll_ctl,epoll_pwait ./d05 sunucu
strace -f -e trace=network,read,write,close ./d05 istemci

cd /kod/05-tcp-buyuk-harf/python
strace -f -e trace=network,read,write python3 tcp-server.py
strace -f -e trace=network,read,write python3 tcp-client.py
```

`./d05`, `ders 05` komutu bir kez çalıştırıldıktan sonra `/calisma` klasöründe
oluşmaktadır. İki programın strace çıktılarını karşılaştırınız: Python bloklayan
soketler kullanmaktadır (`accept4`, `recvfrom` doğrudan beklemektedir); Go ise
bloklamayan soketler kullanmakta, veri yoksa `EAGAIN` alıp `epoll_pwait` ile
beklemektedir.
