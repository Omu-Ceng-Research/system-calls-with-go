# Go ile Sistem Çağrıları — Sistem Programlama

Bu repo, Sistem Programlama dersinde kullanılan sistem çağrılarını küçük ve tek başına
çalıştırılabilen Go programlarıyla göstermektedir. İlk bölüm soketlere ayrılmıştır:
2023-24 dönemindeki soket örneklerinin (Python ve Java) Go karşılıkları ile ödevdeki ağ
ayarlarını gösteren iki örnek yer almaktadır.

Her programda, soket işlemi yapan satırların yanında o satırın ürettiği **sistem
çağrıları** yazılmıştır (`→` ile başlayan yorumlar). Bu yorumlar tahmin değildir;
programların `strace` çıktısından alınmıştır.

> **Ön koşul:** [go-baslangic](https://github.com/Omu-Ceng-Research/go-baslangic) reposundaki
> 15 dersin tamamlanmış olması önerilmektedir. Bu repo aynı Docker ortamını kullanmaktadır.

## İçindekiler

1. [Ortamı başlatmak](#1-ortamı-başlatmak)
2. [Her örnekte ne yapacaksınız?](#2-her-örnekte-ne-yapacaksınız)
3. [Örnekler](#3-örnekler)
4. [strace ile gözlem](#4-strace-ile-gözlem)
5. [Sorun çıkarsa](#5-sorun-çıkarsa)
6. [Klasörde ne var?](#6-klasörde-ne-var)

## 1. Ortamı başlatmak

Docker kurulumu go-baslangic reposundaki [Kurulum](https://github.com/Omu-Ceng-Research/go-baslangic#2-kurulum-bir-kez)
bölümünde anlatılmıştır. Docker kuruluysa bu repoyu indiriniz:

```bash
git clone https://github.com/Omu-Ceng-Research/system-calls-with-go.git
```

Docker Desktop açıkken bu klasörün içinde:

| Sistem | Komut |
|---|---|
| Windows | `baslat.cmd` dosyasına **çift tıklayınız** |
| macOS / Linux | `./baslat.sh` |

İmaj go-baslangic ile aynıdır; daha önce indirildiyse yeniden indirilmemektedir. Bu
ortamın konteyneri go-baslangic'inkinden ayrıdır (`sysprog-cagri`), bu yüzden iki repo
aynı anda açık tutulabilmektedir.

Bazı örnekler iki terminal gerektirmektedir (sunucu ve istemci). Başlatma scriptini bir kez
daha çalıştırınız; script açık olan ortama ikinci bir terminal açmaktadır.

## 2. Her örnekte ne yapacaksınız?

1. Örneğin `main.go` dosyasını baştan sona okuyunuz. Konunun açıklaması dosyanın
   başındadır.
2. `ders NN` ile çalıştırınız (ör. `ders 06 coklu`). Sunucu/istemci örneklerinde
   `ders 05 sunucu` ve ikinci terminalde `ders 05 istemci` yazınız.
3. Açıklamadaki **Gözlemleyiniz** satırlarındaki `strace` komutlarını çalıştırınız ve
   çıktıyı koddaki `→` yorumlarıyla satır satır karşılaştırınız.
4. Dosyanın sonundaki **DENEYİNİZ** ve **SORULAR** bölümlerini yapınız.

```bash
ders 05 sunucu           # 05. örneği derler ve /calisma'da çalıştırır
ders 07 istemci -yazim parcali
cd /calisma              # derlenmiş programlar (d05, d06 ...) buradadır
strace -c -f ./d06 coklu -adet 3    # programın yaptığı sistem çağrılarını sayar
man 2 accept4            # bir sistem çağrısının belgesini gösterir
```

## 3. Örnekler

Numaralar, dersin sistem çağrısı listesindeki sıraya göre verilmiştir. Diğer konuların
(process yönetimi, sinyaller, bellek eşleme, dosyalar) örnekleri ilerleyen haftalarda
01–04 ve 13'ten sonraki numaralarla eklenecektir.

### Ağ ve soketler

| # | Klasör | Konu | Sistem çağrıları | Kaynak |
|---|---|---|---|---|
| 05 | [`05-tcp-buyuk-harf`](05-tcp-buyuk-harf) | TCP sunucusu ve istemcisi, Türkçe büyük harf | `socket`, `bind`, `listen`, `accept4`, `connect` | 2023-24, 1. hafta (Python) |
| 06 | [`06-udp-buyuk-harf`](06-udp-buyuk-harf) | UDP sunucusu ve beş istemci | `recvfrom`, `sendto` | 2023-24, 1. hafta (Python) |
| 07 | [`07-cok-istemcili-tcp`](07-cok-istemcili-tcp) | Çok istemcili sunucu, aynı soketi paylaşan goroutine'ler | `accept4`, `write`, `futex` | 2023-24, 2. hafta (Java) |
| 08 | [`08-nesne-gonderme`](08-nesne-gonderme) | Soket üzerinden nesne: gob, JSON, elle ikili biçim | `write`, `read` | 2023-24, 3. hafta (Java) |
| 09 | [`09-smtp-soket`](09-smtp-soket) | Kütüphanesiz SMTP istemcisi ve sahte SMTP sunucusu | `connect`, `write`, `read` | 2023-24, 3. hafta (telnet) |
| 10 | [`10-smtp-kutuphane`](10-smtp-kutuphane) | Kütüphaneyle renkli ve ek dosyalı e-posta | `connect`, `write` | 2023-24, 4. hafta |
| 11 | [`11-dinleme-ve-ilan-adresi`](11-dinleme-ve-ilan-adresi) | `0.0.0.0`, `127.0.0.1` ve gerçek IP; Docker'da ilan adresi | `bind`, `getsockname`, netlink | Ödev 5.4, T8 |
| 12 | [`12-so-reuseaddr`](12-so-reuseaddr) | TIME_WAIT ve çökme sonrası aynı portu açmak | `setsockopt`, `bind` | Ödev T4, T5 |

### Go'daki bir soket satırı hangi sistem çağrılarını yapar?

Python ya da C'de bir soket fonksiyonu çoğunlukla tek bir sistem çağrısına
karşılık gelmektedir. Go'da ise bir satır birkaç çağrı yapabilmektedir. Örnek 05'te
ayrıntılı olarak gösterilen dört fark şunlardır:

1. **Bloklamayan soketler.** Soketler `SOCK_NONBLOCK` ile açılmaktadır. `accept4` ya da
   `read` o an veri yoksa `EAGAIN` ile hemen dönmektedir; Go goroutine'i park etmekte ve
   soketi `epoll_pwait` ile beklemektedir. strace'teki `= -1 EAGAIN` satırları bir hata
   değil, bu mekanizmanın görüntüsüdür.
2. **epoll kaydı.** Her soket açıldığında `epoll_ctl(EPOLL_CTL_ADD)`, kapatıldığında
   `epoll_ctl(EPOLL_CTL_DEL)` yapılmaktadır.
3. **IPv6 çift yığın.** `":12001"` gibi adressiz bir dinleme, `AF_INET6` soketiyle ve
   `IPV6_V6ONLY=0` ile yapılmaktadır; IPv4 istemcilerin adresi `::ffff:127.0.0.1`
   biçiminde görünmektedir.
4. **Bağlantı ayarları.** Her TCP bağlantısında Go, `TCP_NODELAY` ve keepalive için beş
   `setsockopt` çağırmaktadır.

| Go satırı | Sistem çağrıları |
|---|---|
| `net.Listen("tcp", ":12001")` | `socket`, `setsockopt` × 2, `bind`, `listen`, `epoll_ctl`, `getsockname` |
| `ln.Accept()` | `accept4` (önce `EAGAIN` + `epoll_pwait`), `epoll_ctl`, `getsockname`, `setsockopt` × 5 |
| `net.Dial("tcp", ...)` | `socket`, `connect` (`EINPROGRESS`), `epoll_ctl`, `epoll_pwait`, `getsockopt(SO_ERROR)`, `getpeername`, `getsockname`, `setsockopt` × 5 |
| `conn.Read(b)` | `read` (veri yoksa önce `EAGAIN` + `epoll_pwait`) |
| `conn.Write(b)` | `write` |
| `conn.Close()` | `epoll_ctl(EPOLL_CTL_DEL)`, `close` |
| `net.ListenPacket("udp", ...)` | `socket(SOCK_DGRAM)`, `setsockopt` × 2, `bind` |
| `pc.ReadFrom(b)` / `pc.WriteTo(b, a)` | `recvfrom` / `sendto` |
| `net.Dial("udp", ...)` | `socket(SOCK_DGRAM)`, `setsockopt`, `connect` (paket göndermez), `getsockname` |
| `go f()` | yok (goroutine Go runtime'ının içinde oluşturulmaktadır) |

## 4. strace ile gözlem

Programlar `ders NN` ile derlendiğinde binary `/calisma/dNN` olarak saklanmaktadır.
Gözlem için bu binary `strace` altında doğrudan çalıştırılmalıdır:

```bash
cd /calisma
strace -c -f ./d05 sunucu                            # özet: hangi çağrı kaç kez
strace -f -e trace=network,read,write ./d05 sunucu   # yalnızca belirli çağrılar
strace -f -T -e trace=connect ./d05 istemci          # her çağrının süresi
```

Go runtime birden fazla thread kullandığı için `-f` bayrağı gereklidir. Çıktıdaki
`futex`, `rt_sigaction` ve `epoll_pwait` satırları runtime'a aittir; bunlar `-e trace=...`
ile süzülebilmektedir. Yalnızca ağ çağrılarını görmek için `-e trace=network` kullanılabilmektedir.

## 5. Sorun çıkarsa

| Belirti | Çözüm |
|---|---|
| `bind: address already in use` | Aynı örneğin sunucusu başka bir terminalde açıktır. 07 ve 08 aynı portu (9991) kullanmaktadır; birini kapatınız. |
| `connect: connection refused` | Sunucu çalışmamaktadır ya da farklı bir porttadır. Önce sunucuyu başlatınız. |
| Örnek 09/10 Gmail ile `535` hatası veriyor | Normal Gmail şifresi kabul edilmemektedir. Uygulama şifresi oluşturunuz (örnek 09'un açıklamasına bakınız). |
| Örnek 07'deki bozuk satır sayısı her seferinde farklı | Normaldir; goroutine'lerin sırası her çalıştırmada değişmektedir. |

Diğer sorunlar için go-baslangic reposundaki [Sorun çıkarsa](https://github.com/Omu-Ceng-Research/go-baslangic#8-sorun-çıkarsa)
bölümüne bakınız.

## 6. Klasörde ne var?

| Yol | Açıklama |
|---|---|
| `05-tcp-buyuk-harf` … `12-so-reuseaddr` | Örnek programlar |
| `baslat.cmd`, `baslat.ps1`, `baslat.sh` | Ortamı başlatan scriptler |
| `.devcontainer/` | VS Code "Dev Containers" ayarı |
| `kontrol.sh` | Bütün örnekleri derleyip kısa biçimde sırayla çalıştırmaktadır |

## Lisans

[MIT](LICENSE)
