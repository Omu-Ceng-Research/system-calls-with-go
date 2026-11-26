# 07 — Çok istemcili TCP sunucusu ve aynı soketi paylaşan iş parçacıkları

İstemci tek bir TCP bağlantısı açmakta ve 52 iş parçacığı (Java'da thread, Go'da goroutine)
aynı sokete satır yazmaktadır. Sunucu her bağlantı için ayrı bir iş parçacığı açmaktadır.
2023-24 dönemindeki Java örneği ve Go karşılığı aynı klasörde yer almaktadır. İkisi de
**9991** portunu kullanmaktadır.

| Dosya | Açıklama |
|---|---|
| [`main.go`](main.go) | Go sürümü (sunucu ve istemci tek programda) |
| [`java/Server/Server.java`](java/Server/Server.java) | 2023-24 Java sunucusu |
| [`java/Client/Client.java`](java/Client/Client.java) | 2023-24 Java istemcisi (durdurulana kadar yazar) |

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
| `ders 07 sunucu` | `go run ./07-cok-istemcili-tcp sunucu` |
| `cd /calisma` ve `strace ... ./d07 ...` | `go build -o d07 ./07-cok-istemcili-tcp` ve `strace ... ./d07 ...` |

Linux'ta strace ve nc paketleri kurulu olmalıdır. Windows ve macOS'ta bazı sistem
çağrısı deneyleri (strace) yalnızca Docker'da yapılabilmektedir.

## Go

```bash
# 1. terminal: sunucu (her bağlantı kapanınca satır ve bozuk satır sayısını basar)
ders 07 sunucu

# 2. terminal: üç yazım biçimini sırayla deneyiniz
ders 07 istemci -yazim tek         # her satır tek write: bozuk satır yok
ders 07 istemci -yazim parcali     # her satır iki write: satırlar karışır
ders 07 istemci -yazim kilitli     # iki write, ama ortak kilitle: bozuk satır yok

# goroutine ve satır sayısını değiştirmek
ders 07 istemci -yazim parcali -gorev 100 -tur 50
```

## Java (2023-24)

```bash
# 1. terminal: Server  (aşağıdaki "Java'yı çalıştırmak" bölümüne bakınız)
# 2. terminal: Client  → Ctrl+C ile durdurunuz
```

Java sunucusu her satırı ekrana basmaktadır; istemci durdurulana kadar yazdığı için
çıktı çok hızlı akmaktadır.

### Java'yı çalıştırmak

Docker ortamında Java bulunmamaktadır. İki yol vardır. İkisinde de derlenen `.class` dosyaları repo klasörüne yazılmamaktadır (a) ya da `.gitignore` ile dışarıda tutulmaktadır (b).

**a) Docker ile (Java kurmadan).** Ortam (`baslat.sh`) açıkken, **bilgisayarınızın
kendi terminalinde** (konteynerin içinde değil) repo klasöründe çalıştırınız. Komut, resmî
bir JDK imajını açmakta ve onu Go ortamıyla aynı ağa bağlamaktadır; böylece Java ve Go
programları birbirine `127.0.0.1` üzerinden ulaşabilmektedir. İlk çalıştırmada JDK imajı
(~200 MB) indirilmektedir.

macOS / Linux:

```bash
docker run --rm -it --network container:sysprog-cagri \
  -v "$PWD/07-cok-istemcili-tcp/java/Server":/java -w /java eclipse-temurin:21-jdk \
  bash -c "javac -d /tmp/s *.java && java -cp /tmp/s Server"

docker run --rm -it --network container:sysprog-cagri \
  -v "$PWD/07-cok-istemcili-tcp/java/Client":/java -w /java eclipse-temurin:21-jdk \
  bash -c "javac -d /tmp/c *.java && java -cp /tmp/c Client"
```

Windows (PowerShell):

```powershell
docker run --rm -it --network container:sysprog-cagri -v "${PWD}\07-cok-istemcili-tcp\java\Server:/java" -w /java eclipse-temurin:21-jdk bash -c "javac -d /tmp/s *.java && java -cp /tmp/s Server"

docker run --rm -it --network container:sysprog-cagri -v "${PWD}\07-cok-istemcili-tcp\java\Client:/java" -w /java eclipse-temurin:21-jdk bash -c "javac -d /tmp/c *.java && java -cp /tmp/c Client"
```

**b) Bilgisayarınızda JDK kuruluysa.** Docker'daki programlara bilgisayarınızdan
ulaşılamadığı için bu durumda Go tarafını da bilgisayarınızda (`go run`) çalıştırınız:

```bash
cd 07-cok-istemcili-tcp/java/Server && javac *.java && java Server     # 1. terminal
cd 07-cok-istemcili-tcp/java/Client && javac *.java && java Client     # 2. terminal
go run ./07-cok-istemcili-tcp istemci                                  # repo klasöründe
```

## Karışık: Java ↔ Go

```bash
# Go sunucu + Java istemci: Go sunucusu Java'nın satırlarını da denetlemektedir
ders 07 sunucu                      # konteynerdeki terminal
# Java Client'ı yukarıdaki (a) komutuyla başlatınız, birkaç saniye sonra Ctrl+C
# Go sunucusu: "... kapandı: 118476 satır, 0 bozuk"

# Java sunucu + Go istemci
# Java Server'ı (a) komutuyla başlatınız
ders 07 istemci -gorev 3 -tur 2     # konteynerdeki terminal
```

Java istemcisinde satırlar karışmamaktadır: `println` (autoflush) her satırı tek bir
`write` ile göndermektedir. Go'daki `-yazim parcali` biçimi, bir satırı iki `write`
ile göndererek karışmanın nasıl oluştuğunu göstermektedir.

## Sistem çağrılarını gözlemlemek

```bash
cd /calisma
strace -f -e trace=accept4,read,close ./d07 sunucu
strace -f -e trace=write,futex ./d07 istemci -yazim parcali -gorev 3 -tur 1
strace -f -c -e trace=write,futex,clone ./d07 istemci -yazim kilitli
```

Java tarafı için (bilgisayarınızda JDK ve strace varsa):

```bash
cd 07-cok-istemcili-tcp/java/Client && javac Client.java
strace -f -c -e trace=clone3 java Client      # Ctrl+C: ~82 clone3 (52'si programın thread'leri)
```

Go istemcisinde 52 goroutine için yalnızca birkaç (~6) `clone` görülmektedir; Java
istemcisinde ise her thread bir `clone3`'tür. Go thread'leri `clone`, Java (glibc) ise
`clone3` ile açmaktadır; ikisi de yeni bir işletim sistemi thread'i oluşturmaktadır.
