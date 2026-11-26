// Örnek 05 — TCP soketi: küçük harfli cümleyi büyük harfe çeviren sunucu
//
// 2023-24 dönemindeki "Basic TCP and UDP Sockets with Python" örneğinin
// (tcp-server.py, tcp-client.py) Go karşılığıdır.
//
// İlgili sistem çağrıları: socket, setsockopt, bind, listen, accept4, connect, read,
// write, close, epoll_ctl, epoll_pwait
// Ödevde nerede: 3. bölüm (istemci protokolü), koordinatörün 7000 portu
//
// Çalıştırınız (iki terminal):
//
//	ders 05 sunucu
//	ders 05 istemci                 (ikinci terminalde: baslat.sh / baslat.cmd)
//
// Gözlemleyiniz:
//
//	cd /calisma
//	strace -f -e trace=network,read,write,close,epoll_ctl,epoll_pwait ./d05 sunucu
//	strace -f -e trace=network,read,write,close ./d05 istemci
//
// Python ile Go arasındaki karşılıklar:
//
//	Python                                   Go
//	socket(AF_INET, SOCK_STREAM)             net.Listen("tcp", ...) içinde
//	serverSocket.bind(('', 12001))           ":12001" adresi
//	serverSocket.listen(1)                   net.Listen içinde
//	connectionSocket, addr = accept()        conn, err := ln.Accept()
//	recv(1024)                               conn.Read(tampon[:1024])
//	send(...)                                conn.Write(...)
//	clientSocket.connect((host, port))       net.Dial("tcp", "host:port")
//
// # SATIRLARIN SİSTEM ÇAĞRILARI
//
// Kodun içindeki "→" ile başlayan yorumlar, o satırın yaptığı sistem çağrılarını
// göstermektedir. Bunlar tahmin değildir; programın strace çıktısından alınmıştır.
// Go'nun bir C programından farklı yaptığı dört şey bilinmelidir:
//
//  1. Soketler bloklamayan (SOCK_NONBLOCK) modda açılmaktadır. accept4 ya da read
//     o an veri yoksa beklemek yerine EAGAIN hatasıyla hemen dönmektedir. Go bu
//     durumda goroutine'i park etmekte ve soketi epoll_pwait ile beklemektedir.
//     Veri gelince çağrı tekrarlanmaktadır. strace'teki "= -1 EAGAIN" satırları
//     bir hata değil, bu mekanizmanın görüntüsüdür.
//  2. Her soket açıldığında epoll_ctl(EPOLL_CTL_ADD), kapatıldığında
//     epoll_ctl(EPOLL_CTL_DEL) yapılmaktadır (Go'nun "netpoller"ı).
//  3. ":12001" gibi adressiz bir dinleme, IPv6 soketiyle (AF_INET6) ve
//     IPV6_V6ONLY=0 seçeneğiyle yapılmaktadır. Böylece tek soket hem IPv4 hem IPv6
//     bağlantılarını kabul etmektedir; IPv4 istemcilerin adresi "::ffff:127.0.0.1"
//     biçiminde görünmektedir. Program açılışında Go, makinenin IPv6 desteğini
//     birkaç socket/bind/close çağrısıyla yoklamaktadır.
//  4. Her TCP bağlantısında (hem accept hem connect sonrasında) Go kendiliğinden
//     TCP_NODELAY ve TCP keepalive ayarlarını yapan beş setsockopt çağırmaktadır.
//     Python'da bu ayarlar varsayılan olarak yapılmamaktadır.
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"unicode"
)

const port = "12001"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Kullanım: d05 sunucu | d05 istemci [adres]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "sunucu":
		sunucu()
	case "istemci":
		adres := "127.0.0.1"
		if len(os.Args) > 2 {
			adres = os.Args[2]
		}
		istemci(adres)
	}
}

func sunucu() {
	// ":12001", Python'daki bind(('', 12001)) gibi bütün ağ arayüzlerinin dinlenmesi
	// anlamına gelmektedir.
	//
	// → socket(AF_INET6, SOCK_STREAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_IP) = 3
	// → setsockopt(3, SOL_IPV6, IPV6_V6ONLY, [0])     IPv4 bağlantıları da kabul edilsin
	// → setsockopt(3, SOL_SOCKET, SO_REUSEADDR, [1])  bkz. örnek 12
	// → bind(3, {"::", port 12001})
	// → listen(3, 4096)                                kuyruk boyu /proc/sys/net/core/somaxconn'dan
	// → epoll_ctl(EPOLL_CTL_ADD, 3)                    soket netpoller'a eklenmektedir
	// → getsockname(3)                                 ln.Addr() için gerçek adres öğrenilmektedir
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err) // ör. port kullanımdaysa bind EADDRINUSE döndürmektedir
	}
	fmt.Println("Sunucu almak için hazır, port", port) // → write(1, ...) (1: stdout)

	for {
		// → accept4(3, ..., SOCK_CLOEXEC|SOCK_NONBLOCK) = -1 EAGAIN   henüz bağlanan yok
		// → epoll_pwait(...)                                         bağlantı beklenmektedir
		// → accept4(3, {"::ffff:127.0.0.1", port 51708}, ...) = 4    yeni soket: fd 4
		// → epoll_ctl(EPOLL_CTL_ADD, 4), getsockname(4)
		// → setsockopt(4, TCP_NODELAY), setsockopt(4, SO_KEEPALIVE),
		//   setsockopt(4, TCP_KEEPIDLE), setsockopt(4, TCP_KEEPINTVL), setsockopt(4, TCP_KEEPCNT)
		//
		// accept4 dinleme soketini (fd 3) değiştirmemektedir; her istemci için yeni bir
		// soket (fd 4, 5, ...) döndürmektedir. Dinleme soketi bir sonraki istemciyi beklemektedir.
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		// Python örneğinde istemciler sırayla işlenmekteydi: biri cevap beklerken
		// diğeri accept edilmiyordu. Go'da her bağlantı ayrı bir goroutine'de
		// işlenmektedir (bkz. örnek 07). go anahtar kelimesi sistem çağrısı yapmamaktadır.
		go hizmetEt(conn)
	}
}

func hizmetEt(conn net.Conn) {
	// → epoll_ctl(EPOLL_CTL_DEL, 4) + close(4)     fonksiyon bitince
	defer conn.Close() // Python örneğinde close döngü dışında kaldığı için hiç çağrılmıyordu
	tampon := make([]byte, 1024)

	// → read(4, ..., 1024) = -1 EAGAIN      istemci henüz yazmadı
	// → epoll_pwait(...)                    veri beklenmektedir
	// → read(4, "istanbul", 1024) = 8       gelen kadar bayt dönmektedir, 1024 değil
	n, err := conn.Read(tampon)
	if err != nil {
		log.Print(err) // istemci hiçbir şey yazmadan kapattıysa read 0 döner: io.EOF
		return
	}
	cumle := string(tampon[:n])
	fmt.Printf("%s'den gelen mesaj: %q\n", conn.RemoteAddr(), cumle)

	// strings.ToUpper("istanbul") "ISTANBUL" vermektedir. Türkçe kurallarla
	// büyütmek için unicode.TurkishCase kullanılmalıdır: "İSTANBUL".
	// Bu satır yalnızca bellekte çalışmaktadır; sistem çağrısı yapmamaktadır.
	buyuk := strings.ToUpperSpecial(unicode.TurkishCase, cumle)
	fmt.Printf("İşlem sonucu mesaj: %q\n", buyuk)

	conn.Write([]byte(buyuk)) // → write(4, "İSTANBUL", 9) = 9   ("İ" 2 bayttır)
}

func istemci(adres string) {
	// → socket(AF_INET, SOCK_STREAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_IP) = 3
	// → connect(3, {"127.0.0.1", port 12001}) = -1 EINPROGRESS
	//                     bloklamayan sokette bağlantı arka planda kurulmaktadır
	// → epoll_ctl(EPOLL_CTL_ADD, 3) + epoll_pwait(...)   bağlantının kurulması beklenmektedir
	// → getsockopt(3, SOL_SOCKET, SO_ERROR) = 0          bağlantı başarılı mı?
	// → getpeername(3), getsockname(3)                   karşı ve yerel adres öğrenilmektedir
	// → setsockopt × 5 (TCP_NODELAY ve keepalive)
	conn, err := net.Dial("tcp", net.JoinHostPort(adres, port))
	if err != nil {
		log.Fatal(err) // sunucu açık değilse: "connection refused" (ECONNREFUSED)
	}
	defer conn.Close() // → epoll_ctl(EPOLL_CTL_DEL, 3) + close(3)

	fmt.Print("Küçük harfli cümleler giriniz: ")           // → write(1, ...)
	satir, _ := bufio.NewReader(os.Stdin).ReadString('\n') // → read(0, "istanbul\n", 4096) (0: stdin)
	cumle := strings.TrimRight(satir, "\r\n")
	fmt.Printf("Girilen mesaj: %s\n", cumle)
	fmt.Printf("Encoded mesaj (UTF-8 baytları): % x\n", []byte(cumle))

	conn.Write([]byte(cumle)) // → write(3, "istanbul", 8) = 8
	tampon := make([]byte, 1024)

	// → read(3, ..., 1024) = -1 EAGAIN, epoll_pwait(...), read(3, "İSTANBUL", 1024) = 9
	n, err := conn.Read(tampon)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sunucudan gelen mesaj: %s\n", tampon[:n])
}

/*
DENEYİNİZ

 1. Yukarıdaki iki strace komutunu çalıştırınız ve koddaki "→" yorumlarıyla satır satır
    karşılaştırınız. accept4 ve read'in önce EAGAIN döndürdüğünü bulunuz.
 2. "ığüşöç istanbul" cümlesini gönderiniz. Encoded çıktısında Türkçe harflerin
    kaç bayt tuttuğunu sayınız. 1024 baytlık tampona kaç Türkçe harf sığmaktadır?
 3. hizmetEt içinde strings.ToUpperSpecial yerine strings.ToUpper kullanınız ve
    "istanbul" cümlesinin sonucunu karşılaştırınız.
 4. Sunucu kapalıyken istemciyi strace altında çalıştırınız. connect hangi hatayı
    döndürmektedir?
 5. Sunucuya nc ile bağlanınız:  echo "merhaba dünya" | nc -q1 127.0.0.1 12001
    nc de strace altında çalıştırılabilir: C ile yazılmış bir programın çağrıları
    Go'nunkilerden nasıl farklıdır? (EAGAIN ve epoll var mıdır?)
 6. 2000 baytlık bir cümle gönderiniz:
    python3 -c "print('a'*2000)" | nc -q1 127.0.0.1 12001 | wc -c

SORULAR

  - TCP bir bayt akışıdır, mesaj sınırı yoktur. Tek bir Read(1024) çağrısının
    istemcinin gönderdiği cümlenin tamamını döndüreceğinin garantisi var mıdır?
  - Ödevdeki protokolde komutların "\n" ile bitmesi bu sorunu nasıl çözmektedir?
  - Sunucu ":12001" yerine "127.0.0.1:12001" dinleseydi başka bir makineden
    bağlanılabilir miydi? strace'te socket çağrısı hangi adres ailesiyle yapılırdı?
    (bkz. örnek 11)
*/
