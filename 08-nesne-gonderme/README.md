# 08 — Soket üzerinden nesne göndermek

İstemci bir kişi nesnesini (ad, yaş, boy) serileştirip sokete yazmakta, sunucu okuyup
alanlarını basmaktadır. 2023-24 dönemindeki Java örneği (`ObjectOutputStream`) ve Go
karşılığı (gob, JSON ve elle ikili biçim) aynı klasörde yer almaktadır. İkisi de **9991**
portunu kullanmaktadır.

| Dosya | Açıklama |
|---|---|
| [`main.go`](main.go) | Go sürümü (sunucu ve istemci tek programda, üç biçim) |
| [`java/Server/`](java/Server) | 2023-24 Java sunucusu (`Server.java`, `Person.java`) |
| [`java/Client/`](java/Client) | 2023-24 Java istemcisi (`Client.java`, `Person.java`) |

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
| `ders 08 sunucu` | `go run ./08-nesne-gonderme sunucu` |
| `cd /calisma` ve `strace ... ./d08 ...` | `go build -o d08 ./08-nesne-gonderme` ve `strace ... ./d08 ...` |

Linux'ta strace ve nc paketleri kurulu olmalıdır. Windows ve macOS'ta bazı sistem
çağrısı deneyleri (strace) yalnızca Docker'da yapılabilmektedir.

## Go

```bash
# 1. terminal: sunucu (bağlantıları sırayla kabul eder)
ders 08 sunucu

# 2. terminal: istemci, üç biçim
ders 08 istemci -bicim gob
ders 08 istemci -bicim json
ders 08 istemci -bicim ikili

# Sunucu yerine nc ile dinleyip giden baytları görmek
nc -l -p 9991 | hexdump -C                # 1. terminal
ders 08 istemci -bicim json         # 2. terminal
```

## Java (2023-24)

```bash
# 1. terminal: Server
# 2. terminal: Client  (aşağıdaki "Java'yı çalıştırmak" bölümüne bakınız)
```

Java sunucusu ilk nesneden sonra dinleme soketini de kapattığı için
`SocketException: Socket is closed` hatasıyla sona ermektedir; her deneme için sunucu
yeniden başlatılmalıdır.

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
  -v "$PWD/08-nesne-gonderme/java/Server":/java -w /java eclipse-temurin:21-jdk \
  bash -c "javac -d /tmp/s *.java && java -cp /tmp/s Server"

docker run --rm -it --network container:sysprog-cagri \
  -v "$PWD/08-nesne-gonderme/java/Client":/java -w /java eclipse-temurin:21-jdk \
  bash -c "javac -d /tmp/c *.java && java -cp /tmp/c Client"
```

Windows (PowerShell):

```powershell
docker run --rm -it --network container:sysprog-cagri -v "${PWD}\08-nesne-gonderme\java\Server:/java" -w /java eclipse-temurin:21-jdk bash -c "javac -d /tmp/s *.java && java -cp /tmp/s Server"

docker run --rm -it --network container:sysprog-cagri -v "${PWD}\08-nesne-gonderme\java\Client:/java" -w /java eclipse-temurin:21-jdk bash -c "javac -d /tmp/c *.java && java -cp /tmp/c Client"
```

**b) Bilgisayarınızda JDK kuruluysa.** Docker'daki programlara bilgisayarınızdan
ulaşılamadığı için bu durumda Go tarafını da bilgisayarınızda (`go run`) çalıştırınız:

```bash
cd 08-nesne-gonderme/java/Server && javac *.java && java Server     # 1. terminal
cd 08-nesne-gonderme/java/Client && javac *.java && java Client     # 2. terminal
go run ./08-nesne-gonderme istemci                                  # repo klasöründe
```

## Java ile Go neden konuşamaz?

Java'nın `ObjectOutputStream` biçimi yalnızca Java'ya, Go'nun gob biçimi yalnızca Go'ya
özgüdür. Go istemcisi Java sunucusuna JSON gönderirse Java sunucusu şu hatayı vermektedir:

```
java.io.StreamCorruptedException: invalid stream header: 6A7B2241
```

`6A7B2241`, gelen ilk dört baytın onaltılık hâlidir: `j{"A`. Java, akışın
`AC ED 00 05` (sihirli sayı ve sürüm) ile başlamasını beklemektedir. Farklı diller
arasında nesne göndermek için JSON ya da protobuf gibi dilden bağımsız bir biçim
gerekmektedir; ödevde bu yüzden gRPC/protobuf kullanılmaktadır.

## Sistem çağrılarını gözlemlemek

```bash
cd /calisma
strace -f -e trace=read ./d08 sunucu
strace -f -e trace=write ./d08 istemci -bicim gob
```

Java tarafı için (bilgisayarınızda JDK ve strace varsa):

```bash
cd 08-nesne-gonderme/java/Server && javac *.java
strace -f -e trace=read java Server 2>&1 | grep -A30 'accept'
```

Java sunucusu 86 baytlık nesneyi 24 ayrı `read` ile okumaktadır; Go sunucusu ise
`bufio.Reader` sayesinde tek `read` ile okumaktadır.
