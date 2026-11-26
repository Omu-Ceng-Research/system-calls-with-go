// Örnek 07 — Çok istemcili TCP sunucusu ve aynı soketi paylaşan goroutine'ler
//
// 2023-24 dönemindeki "Multi-threaded TCP Sockets with Java" örneğinin
// (Server.java, Client.java) Go karşılığıdır.
//
// İlgili sistem çağrıları: accept4, read, write, clone, futex, epoll_pwait
// Ödevde nerede: koordinatörün aynı anda birden fazla istemciye hizmet vermesi (3. bölüm),
// düğümlere paralel gRPC istekleri (5.4)
//
// Çalıştırınız (iki terminal):
//
//	ders 07 sunucu
//	ders 07 istemci -yazim tek        (her satır tek Write: satırlar bozulmamaktadır)
//	ders 07 istemci -yazim parcali    (her satır iki Write: satırlar karışmaktadır)
//	ders 07 istemci -yazim kilitli    (iki Write, ama ortak bir kilitle)
//
// Gözlemleyiniz:
//
//	cd /calisma
//	strace -f -e trace=accept4,read,close ./d07 sunucu
//	strace -f -e trace=write,futex ./d07 istemci -yazim parcali -gorev 3 -tur 1
//	strace -f -c -e trace=write,futex ./d07 istemci -yazim kilitli
//
// Kodun içindeki "→" ile başlayan yorumlar, o satırın yaptığı sistem çağrılarını
// göstermektedir ve strace çıktısından alınmıştır. Go'ya özgü EAGAIN, epoll ve IPv6
// ayrıntıları örnek 05'te açıklanmıştır.
//
// Java örneğindeki yapı korunmuştur: istemci tek bir soket açmakta ve 52 iş
// parçacığı (burada goroutine) aynı sokete yazmaktadır. Sunucu her bağlantı için
// ayrı bir goroutine açmaktadır (Java'daki new Thread(new ClientHandler(...))).
//
// Java örneğindeki iki noktaya dikkat edilmelidir:
//
//  1. ReentrantLock her ClientHandler nesnesinde ayrı oluşturulmaktaydı. Her iş
//     parçacığının kendi kilidi olduğu için kilit hiçbir şeyi korumuyordu. Bir kilidin
//     işe yaraması için onu paylaşan herkesin aynı kilidi kullanması gerekmektedir.
//  2. Birden fazla iş parçacığı aynı sokete yazdığında satırların karışıp karışmaması,
//     bir satırın tek bir write(2) ile mi yoksa birkaç parça hâlinde mi gönderildiğine
//     bağlıdır. Go'da tek bir conn.Write çağrısı bütün olarak yazılmaktadır; ancak
//     bir satır iki Write ile gönderilirse araya başka goroutine'in verisi girebilmektedir.
package main

import (
	"bufio"   // NewScanner
	"flag"    // NewFlagSet, ExitOnError
	"fmt"     // Fprintln, Println, Printf, Sscanf, Sprintf, Sprint
	"log"     // Fatal, Print
	"net"     // Listen, Conn, Dial, JoinHostPort
	"os"      // Args, Stderr, Exit
	"strings" // Repeat
	"sync"    // Mutex, WaitGroup
)

const port = "9991"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Kullanım: d07 sunucu | d07 istemci [-yazim tek|parcali|kilitli] [-gorev 52] [-tur 10]")
		os.Exit(1)
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	yazim := fs.String("yazim", "tek", "satırın sokete yazılma biçimi: tek, parcali, kilitli")
	gorev := fs.Int("gorev", 52, "aynı soketi paylaşan goroutine sayısı")
	tur := fs.Int("tur", 10, "her goroutine'in göndereceği satır sayısı")
	adres := fs.String("adres", "127.0.0.1", "sunucunun adresi")
	fs.Parse(os.Args[2:])

	switch os.Args[1] {
	case "sunucu":
		sunucu()
	case "istemci":
		istemci(*adres, *yazim, *gorev, *tur)
	}
}

func sunucu() {
	// → socket(AF_INET6, SOCK_STREAM, ...) + setsockopt × 2 + bind + listen (bkz. örnek 05)
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Merhaba, port", port)
	for {
		// → accept4(3, ...) = 4   (ikinci istemci için = 5, üçüncü için = 6 ...)
		// → setsockopt × 5 (TCP_NODELAY ve keepalive)
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		// Java: new Thread(new ClientHandler(clientSocket)).start()
		// Java'da new Thread bir clone3(2) yapıp yeni bir işletim sistemi thread'i
		// açmaktadır. go ifadesi ise sistem çağrısı yapmamaktadır; goroutine Go runtime'ının
		// içinde oluşturulmaktadır. Runtime gerekirse birkaç thread'i clone(2) ile kendisi
		// açmaktadır: 52 goroutine'li istemcide strace ~6 clone göstermektedir.
		go istemciyleIlgilen(conn)
	}
}

// istemciyleIlgilen, bağlantı kapanana kadar satır okumakta ve her satırın
// bozulmadan gelip gelmediğini denetlemektedir.
func istemciyleIlgilen(conn net.Conn) {
	defer conn.Close() // → epoll_ctl(EPOLL_CTL_DEL, 4) + close(4)
	okuyucu := bufio.NewScanner(conn)
	okuyucu.Buffer(make([]byte, 64*1024), 1<<20)
	var toplam, bozuk int
	// Scan, her satır için ayrı bir read yapmamaktadır:
	// → read(4, ..., 65536) = 1015      gelen kadar bayt (bir ya da birkaç satır)
	// → read(4, ..., 65536) = -1 EAGAIN + epoll_pwait(...)   tampon boşalınca
	// → read(4, "", 65536) = 0          istemci bağlantıyı kapattı: döngü biter
	// 520 satır için strace'te ~800 read görülmektedir; satır sayısıyla eşit değildir.
	for okuyucu.Scan() {
		satir := okuyucu.Text()
		toplam++
		if !saglamMi(satir) {
			bozuk++
			if bozuk <= 3 {
				fmt.Printf("BOZUK SATIR: %.70s...\n", satir)
			}
		}
	}
	fmt.Printf("%s kapandı: %d satır, %d bozuk\n", conn.RemoteAddr(), toplam, bozuk)
}

// Satır biçimi: "gorev=<n> mesaj <n tekrar eden>". Başka bir goroutine'in verisi
// araya girdiyse tekrar eden kısım gorev numarasıyla uyuşmamaktadır.
// Java istemcisinin (java/Client) biçimi de tanınmaktadır:
// "Thread_id: <id> Message <id, 1000 kez tekrar>".
func saglamMi(satir string) bool {
	var no int
	if _, err := fmt.Sscanf(satir, "Thread_id: %d Message ", &no); err == nil {
		return satir == fmt.Sprintf("Thread_id: %d Message %s", no, strings.Repeat(fmt.Sprint(no), 1000))
	}
	if _, err := fmt.Sscanf(satir, "gorev=%d mesaj ", &no); err != nil {
		return false
	}
	beklenen := fmt.Sprintf("gorev=%d mesaj %s", no, strings.Repeat(fmt.Sprint(no%10), 1000))
	return satir == beklenen
}

func istemci(adres, yazim string, gorev, tur int) {
	// → socket + connect = -1 EINPROGRESS + epoll_pwait + getsockopt(SO_ERROR) ... (bkz. örnek 05)
	conn, err := net.Dial("tcp", net.JoinHostPort(adres, port))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Bağlanıldı!")

	// Bütün goroutine'lerin paylaştığı TEK kilit. Kilit boştayken Lock/Unlock sistem
	// çağrısı yapmamaktadır (yalnızca atomik bir bellek işlemi). Kilit doluysa ve bekleme
	// uzarsa goroutine'i taşıyan thread futex(2) ile uyutulmakta ve uyandırılmaktadır.
	var kilit sync.Mutex
	var wg sync.WaitGroup
	for g := 0; g < gorev; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			bas := fmt.Sprintf("gorev=%d mesaj ", g)
			govde := strings.Repeat(fmt.Sprint(g%10), 1000) + "\n"
			for i := 0; i < tur; i++ {
				switch yazim {
				case "tek":
					// → write(3, "gorev=2 mesaj 2222...\n", 1015) = 1015
					// Satırın tamamı tek write(2). Go, aynı sokete yapılan Write
					// çağrılarını kendi içinde sıraya koymaktadır; iki write iç içe geçmemektedir.
					conn.Write([]byte(bas + govde))
				case "parcali":
					// → write(3, "gorev=2 mesaj ", 14)
					// → write(3, "2222...\n", 1001)
					// İki ayrı write(2). İkisinin arasına başka bir goroutine'in write'ı
					// girebilmektedir. strace -f çıktısında farklı thread'lerin write'ları
					// iç içe görülmektedir.
					conn.Write([]byte(bas))
					conn.Write([]byte(govde))
				case "kilitli":
					kilit.Lock()            // → çekişme varsa futex(FUTEX_WAIT)
					conn.Write([]byte(bas)) // → write(3, "gorev=2 mesaj ", 14)
					conn.Write([]byte(govde))
					kilit.Unlock() // → bekleyen varsa futex(FUTEX_WAKE)
				}
			}
		}(g)
	}
	wg.Wait()
	fmt.Printf("%d goroutine × %d satır gönderildi (%s)\n", gorev, tur, yazim)
	conn.Close() // → close(3): sunucunun read'i 0 döndürmektedir (EOF)
}

/*
DENEYİNİZ

 1. Üç yazım biçimini sırayla deneyiniz ve sunucunun bildirdiği bozuk satır
    sayılarını karşılaştırınız. "parcali" biçimi her çalıştırmada aynı sonucu
    vermekte midir?
 2. "kilitli" biçiminde kilidi goroutine'in içinde tanımlayınız (her goroutine'e ayrı
    bir kilit). Sonuç Java örneğindeki ReentrantLock hatasıyla aynı mıdır?
 3. go build -race -o d07r ./07-cok-istemcili-tcp ile derleyip "parcali" biçimini
    çalıştırınız. Race detector bir şey bildirmekte midir? Neden?
 4. Aynı anda iki istemci başlatınız. Sunucu iki bağlantıyı ayrı ayrı saymaktadır.

SORULAR

  - Java örneğindeki istemci sonsuz döngüyle yazmaktadır ve sunucu yavaş okursa ne
    olmaktadır? (ipucu: soketin gönderme tamponu dolunca write(2) bloklanmaktadır)
  - Ödevde koordinatör bir düğüme aynı gRPC bağlantısı üzerinden birden fazla
    goroutine ile istek göndermektedir. gRPC bu karışmayı nasıl önlemektedir?
*/
