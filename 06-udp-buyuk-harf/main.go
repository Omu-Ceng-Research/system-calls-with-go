// Örnek 06 — UDP soketi: bağlantısız sunucu ve aynı anda çalışan beş istemci
//
// 2023-24 dönemindeki "Basic TCP and UDP Sockets with Python" örneğinin
// (UDP Socket/server.py, client.py … client5.py) Go karşılığıdır. Python'daki beş
// ayrı istemci dosyası burada tek programın argümanı olmuştur.
//
// İlgili sistem çağrıları: socket(SOCK_DGRAM), bind, recvfrom, sendto
// Ödevde nerede: Ödevde UDP kullanılmamaktadır; TCP ile farkın anlaşılması içindir.
//
// Çalıştırınız (iki terminal):
//
//	ders 06 sunucu
//	ders 06 istemci a               (b, c, d, e harfleriyle başka terminallerde de)
//	ders 06 coklu                   (beş istemci tek programda, aynı anda)
//
// Gözlemleyiniz:
//
//	cd /calisma
//	strace -f -e trace=network,read,write,epoll_pwait ./d06 sunucu
//	strace -f -e trace=network,read,write,epoll_pwait ./d06 istemci a -adet 2
//
// Kodun içindeki "→" ile başlayan yorumlar, o satırın yaptığı sistem çağrılarını
// göstermektedir ve strace çıktısından alınmıştır. Go'ya özgü EAGAIN, epoll ve IPv6
// ayrıntıları örnek 05'te açıklanmıştır.
//
// TCP'den farklar:
//
//   - listen ve accept yoktur. Sunucu tek bir soketle bütün istemcilere cevap
//     vermektedir; her paketin kimden geldiği recvfrom(2) ile öğrenilmektedir.
//   - Mesaj sınırları korunmaktadır: bir sendto bir paket, bir recvfrom bir pakettir.
//   - Teslim garantisi yoktur. Kaybolan paket için hata alınmamaktadır; bu yüzden
//     istemci cevabı sonsuza kadar değil, bir zaman aşımıyla beklemelidir.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

const port = "12000"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Kullanım: d06 sunucu | d06 istemci <harf> [-adet N] | d06 coklu [-adet N]")
		os.Exit(1)
	}
	mod := os.Args[1]
	harf := "a"
	argumanlar := os.Args[2:]
	if mod == "istemci" && len(argumanlar) > 0 && !strings.HasPrefix(argumanlar[0], "-") {
		harf, argumanlar = argumanlar[0], argumanlar[1:]
	}
	fs := flag.NewFlagSet(mod, flag.ExitOnError)
	adet := fs.Int("adet", 0, "gönderilecek mesaj sayısı (0: durdurulana kadar)")
	adres := fs.String("adres", "127.0.0.1", "sunucunun adresi")
	fs.Parse(argumanlar)

	switch mod {
	case "sunucu":
		sunucu()
	case "istemci":
		istemci(*adres, harf, *adet)
	case "coklu":
		var wg sync.WaitGroup
		for _, h := range []string{"a", "b", "c", "d", "e"} {
			wg.Add(1)
			go func(h string) { defer wg.Done(); istemci(*adres, h, *adet) }(h)
		}
		wg.Wait()
	}
}

func sunucu() {
	// → socket(AF_INET6, SOCK_DGRAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_IP) = 3
	// → setsockopt(3, SOL_IPV6, IPV6_V6ONLY, [0])
	// → setsockopt(3, SOL_SOCKET, SO_BROADCAST, [1])
	// → bind(3, {"::", port 12000})
	// listen ve accept yoktur: UDP'de "bağlantı" kavramı bulunmamaktadır.
	pc, err := net.ListenPacket("udp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sunucu almak için hazır, port", port)

	tampon := make([]byte, 2048)
	for {
		// → recvfrom(3, ..., 2048) = -1 EAGAIN, epoll_pwait(...)    paket beklenmektedir
		// → recvfrom(3, "aaaaaaaaaa", 2048, 0, {"::ffff:127.0.0.1", port 56786}) = 10
		// recvfrom, veriyle birlikte gönderenin adresini de döndürmektedir.
		n, istemciAdresi, err := pc.ReadFrom(tampon)
		if err != nil {
			log.Print(err)
			continue
		}
		buyuk := strings.ToUpperSpecial(unicode.TurkishCase, string(tampon[:n]))
		// → sendto(3, "AAAAAAAAAA", 10, 0, {"::ffff:127.0.0.1", port 56786}) = 10
		// Her gönderimde hedef adres ayrıca verilmektedir; aynı soket herkese yazmaktadır.
		pc.WriteTo([]byte(buyuk), istemciAdresi)
		fmt.Printf("%-21s %s\n", istemciAdresi, buyuk)
	}
}

func istemci(adres, harf string, adet int) {
	// → socket(AF_INET, SOCK_DGRAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_IP) = 3
	// → setsockopt(3, SOL_SOCKET, SO_BROADCAST, [1])
	// → connect(3, {"127.0.0.1", port 12000}) = 0
	// → getsockname(3) = {"127.0.0.1", port 48595}    çekirdeğin seçtiği yerel port
	//
	// UDP'de connect(2) bağlantı kurmamaktadır ve ağa hiçbir paket göndermemektedir;
	// yalnızca soketin varsayılan hedefini belirlemektedir. Bu yüzden TCP'deki gibi
	// EINPROGRESS beklenmemekte, hemen 0 dönmektedir. Sunucu kapalı olsa bile başarılıdır.
	conn, err := net.Dial("udp", net.JoinHostPort(adres, port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close() // → close(3)

	mesaj := strings.Repeat(harf, 10)
	tampon := make([]byte, 2048)
	for i := 0; adet == 0 || i < adet; i++ {
		// Soket connect edildiği için sendto yerine write kullanılabilmektedir.
		conn.Write([]byte(mesaj)) // → write(3, "aaaaaaaaaa", 10) = 10
		// Python örneğindeki recvfrom zaman aşımı olmadan beklemekteydi; sunucu
		// kapalıysa istemci sonsuza kadar asılı kalırdı.
		// SetReadDeadline sistem çağrısı yapmamaktadır: süre Go runtime'ının
		// zamanlayıcısında tutulmakta ve epoll_pwait'in bekleme süresine yansımaktadır.
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		// → read(3, ..., 2048) = -1 EAGAIN, epoll_pwait(...), read(3, "AAAAAAAAAA", 2048) = 10
		n, err := conn.Read(tampon)
		if err != nil {
			fmt.Printf("[%s] cevap yok: %v\n", harf, err)
		} else {
			fmt.Printf("[%s] %s\n", harf, tampon[:n])
		}
		time.Sleep(time.Second)
	}
}

/*
DENEYİNİZ

 1. Sunucu çalışırken üç ayrı terminalde a, b ve c istemcilerini başlatınız.
    Sunucu çıktısında her istemcinin farklı bir kaynak portu olduğunu gözlemleyiniz.
 2. İstemciler çalışırken sunucuyu Ctrl+C ile durdurunuz ve yeniden başlatınız.
    İstemciler ne yapmaktadır? Aynı deney TCP'de (örnek 05) nasıl sonuçlanırdı?
 3. Sunucu kapalıyken istemci çalıştırıldığında hata "i/o timeout" mu yoksa
    "connection refused" mı olmaktadır? Neden? (ipucu: ICMP port unreachable)
 4. strace çıktısında sunucunun hiç accept çağrısı yapmadığını doğrulayınız.

SORULAR

  - 2048 bayttan büyük bir UDP paketi gelirse fazlası ne olmaktadır?
  - Ödevde kalp atışları (her 2 saniyede bir) UDP ile gönderilebilir miydi?
    Kaybolan bir kalp atışı ne zaman sorun olurdu?
*/
