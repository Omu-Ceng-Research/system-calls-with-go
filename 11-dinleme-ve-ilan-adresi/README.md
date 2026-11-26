# 11 — Dinleme adresi ve ilan adresi

Sunucunun hangi adresi dinlediği (`127.0.0.1`, gerçek IP ya da `0.0.0.0`), ona
kimlerin bağlanabileceğini belirlemektedir. Program makinenin ağ arayüzlerini listelemekte,
verilen adresi dinlemekte ve her arayüzün IP'si üzerinden kendine bağlanmayı denemektedir.
Ödevdeki `-dinle` / `-ilan` ayrımının (5.4) ve laboratuvar demosunun (T8) temelidir.
Bu konu için eski yıllardan bir örnek bulunmamaktadır.

| Dosya | Açıklama |
|---|---|
| [`main.go`](main.go) | Go programı |

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
| `ders 11` | `go run ./11-dinleme-ve-ilan-adresi` |
| `cd /calisma` ve `strace ... ./d11 ...` | `go build -o d11 ./11-dinleme-ve-ilan-adresi` ve `strace ... ./d11 ...` |

Linux'ta strace ve nc paketleri kurulu olmalıdır. Windows ve macOS'ta bazı sistem
çağrısı deneyleri (strace) yalnızca Docker'da yapılabilmektedir.

## Çalıştırmak

Tek terminal yeterlidir:

```bash
ders 11                              # yalnızca 127.0.0.1:7100
ders 11 -dinle 0.0.0.0:7100          # bütün arayüzler
ders 11 -dinle 172.17.0.2:7100       # yalnızca bu IP: ders 11 çıktısındaki 127.0.0.1 dışındaki
                                     # bir adresi yazınız (Docker'da genellikle eth0, 172.17.x.x)
```

İki makineyle (laboratuvarda):

```bash
# 1. makine (Docker kullanmadan, Go kuruluysa)
go run ./11-dinleme-ve-ilan-adresi -dinle 0.0.0.0:7100

# 2. makine: 1. makinenin IP'sine bağlanmayı deneyiniz
nc -vz <1. makinenin IP'si> 7100
```

Docker içinde çalışan bir programa başka bir makineden ulaşılamamaktadır; program bunu
`UYARI` satırıyla bildirmektedir (ayrıntı `main.go` dosyasının başında).

## Sistem çağrılarını gözlemlemek

```bash
cd /calisma
strace -f -e trace=network,close ./d11
```

Arayüz listesinin bir dosyadan değil, `AF_NETLINK` soketiyle çekirdeğe sorularak
alındığını; `bind` çağrısına verilen adresin dinleme adresiyle aynı olduğunu doğrulayınız.
