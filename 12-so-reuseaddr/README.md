# 12 — TIME_WAIT ve SO_REUSEADDR

Sunucu kapandıktan hemen sonra aynı portu yeniden açmaya çalışmaktadır. `SO_REUSEADDR`
kapalıyken `bind` çağrısı `address already in use` hatası vermekte, açıkken
başarılı olmaktadır. Ödevdeki çökme testlerinde (T4, T5) düğüm ve koordinatör `kill -9`
sonrası hemen yeniden başlatılmaktadır; bu örnek, bunun neden mümkün olduğunu
göstermektedir. Bu konu için eski yıllardan bir örnek bulunmamaktadır.

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
| `ders 12` | `go run ./12-so-reuseaddr` |
| `cd /calisma` ve `strace ... ./d12 ...` | `go build -o d12 ./12-so-reuseaddr` ve `strace ... ./d12 ...` |

Linux'ta strace ve nc paketleri kurulu olmalıdır. Windows ve macOS'ta bazı sistem
çağrısı deneyleri (strace) yalnızca Docker'da yapılabilmektedir.

## Çalıştırmak

Tek terminal yeterlidir:

```bash
ders 12
ders 12        # hemen ikinci kez: TIME_WAIT'teki bağlantı sayısı artmaktadır
```

TIME_WAIT durumundaki bağlantıları görmek için (durum kodu `06`):

```bash
cat /proc/net/tcp
```

## Aynı sorun Python'da

Python soketleri `SO_REUSEADDR` açmadan oluşturmaktadır. Örnek 05'in Go sunucusunu
çalıştırıp bir istemci bağladıktan sonra kapatınız ve hemen Python sunucusunu
başlatınız:

```bash
python3 /kod/05-tcp-buyuk-harf/python/tcp-server.py
# OSError: [Errno 98] Address already in use
```

## Sistem çağrılarını gözlemlemek

```bash
cd /calisma
strace -f -e trace=socket,setsockopt,bind,listen ./d12
```

Çıktıda Go'nun her dinleme soketinde `SO_REUSEADDR`'ı kendiliğinden açtığını
(`[1]`), programın ise bir denemede onu kapattığını (`[0]`) ve ardından `bind`
çağrısının `EADDRINUSE` döndürdüğünü bulunuz.
