// Örnek 12 — TIME_WAIT ve SO_REUSEADDR: çöken sunucu neden aynı portu hemen açamaz?
//
// İlgili sistem çağrıları: setsockopt(SO_REUSEADDR), bind, close
// Ödevde nerede: T4 ve T5 (kill -9 sonrası düğümün ve koordinatörün aynı portla
// hemen yeniden başlaması)
//
// Çalıştırınız:
//
//	ders 12
//
// Gözlemleyiniz:
//
//	cd /calisma
//	strace -f -e trace=socket,setsockopt,bind,listen ./d12
//
// Kodun içindeki "→" ile başlayan yorumlar, o satırın yaptığı sistem çağrılarını
// göstermektedir ve strace çıktısından alınmıştır.
//
// Bir TCP bağlantısını önce kapatan taraf, bağlantıyı hemen unutmamaktadır.
// Bağlantı yaklaşık 60 saniye TIME_WAIT durumunda kalmaktadır; böylece ağda
// gecikmiş eski paketler yeni bir bağlantıya karışmamaktadır.
//
// Sunucu çöküp yeniden başladığında eski bağlantılar hâlâ TIME_WAIT'te
// olabilmektedir. SO_REUSEADDR seçeneği kapalıysa bind(2) bu durumda
// EADDRINUSE ("address already in use") hatası vermektedir ve sunucu 60 saniye
// boyunca açılamamaktadır.
//
// Go, net.Listen ile açılan her dinleme soketinde SO_REUSEADDR'ı kendiliğinden
// açmaktadır (strace çıktısında görülmektedir). Bu örnek, seçenek kapatıldığında ne
// olduğunu göstermektedir. C ile yazılmış bir sunucuda bu seçeneğin elle açılması
// gerekmektedir.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

const adres = "127.0.0.1:7200"

func main() {
	// 1. Sunucu açılmakta, bir istemci bağlanmakta ve sunucu bağlantıyı önce kapatmaktadır.
	// → socket(AF_INET, SOCK_STREAM, ...) = 3
	// → setsockopt(3, SOL_SOCKET, SO_REUSEADDR, [1])   Go bunu her dinleme soketinde yapmaktadır
	// → bind(3, {"127.0.0.1", port 7200}), listen(3, 4096)
	ln, err := net.Listen("tcp", adres)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		c, err := ln.Accept()
		if err == nil {
			c.Write([]byte("merhaba\n"))
			// → close(4): çekirdek karşı tarafa FIN göndermektedir. Önce kapatan taraf
			// sunucu olduğu için TIME_WAIT sunucu tarafında kalmaktadır.
			c.Close()
		}
	}()
	c, err := net.Dial("tcp", adres)
	if err != nil {
		log.Fatal(err)
	}
	tampon := make([]byte, 64)
	c.Read(tampon)
	time.Sleep(100 * time.Millisecond)
	c.Close()
	ln.Close() // → close(3): sunucunun "çöktüğü" an. Port boşalmış görünmektedir, ama...
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("Sunucu kapandı. 7200 portunda TIME_WAIT'te bekleyen bağlantı: %d\n\n", timeWaitSay(7200))

	// 2. Aynı port SO_REUSEADDR kapalıyken açılmaya çalışılmaktadır.
	// → socket(AF_INET, ...) = 3
	// → setsockopt(3, SO_REUSEADDR, [1])     Go'nun kendi ayarı
	// → setsockopt(3, SO_REUSEADDR, [0])     Control fonksiyonumuz bunu geri almaktadır
	// → bind(3, {"127.0.0.1", port 7200}) = -1 EADDRINUSE
	_, err = dinle(false)
	fmt.Printf("SO_REUSEADDR kapalı:  %v\n", sonuc(err))

	// 3. Aynı port SO_REUSEADDR açıkken (Go'nun varsayılanı) açılmaktadır.
	// → setsockopt(3, SO_REUSEADDR, [1]) × 2, bind(3, ...) = 0
	ln2, err := dinle(true)
	fmt.Printf("SO_REUSEADDR açık:    %v\n", sonuc(err))
	if ln2 != nil {
		ln2.Close()
	}
}

// dinle, soket bind edilmeden hemen önce SO_REUSEADDR seçeneğini ayarlamaktadır.
// net.ListenConfig.Control, Go'nun soketi oluşturduğu ile bind ettiği an arasında
// çağrılmaktadır.
func dinle(reuse bool) (net.Listener, error) {
	deger := 0
	if reuse {
		deger = 1
	}
	lc := net.ListenConfig{
		Control: func(network, address string, rc syscall.RawConn) error {
			var serr error
			rc.Control(func(fd uintptr) {
				// → setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, [deger], 4)
				serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, deger)
			})
			return serr
		},
	}
	return lc.Listen(context.Background(), "tcp", adres)
}

func sonuc(err error) string {
	if err != nil {
		return "AÇILAMADI → " + err.Error()
	}
	return "açıldı"
}

// timeWaitSay, /proc/net/tcp dosyasını okuyarak verilen yerel porttaki TIME_WAIT
// (durum kodu 06) bağlantılarını saymaktadır. ss ve netstat da bu dosyayı okumaktadır.
func timeWaitSay(port int) int {
	veri, err := os.ReadFile("/proc/net/tcp") // → openat + read + close: çekirdeğin TCP tablosu
	if err != nil {
		return -1
	}
	arananPort := fmt.Sprintf(":%04X", port)
	say := 0
	for _, satir := range strings.Split(string(veri), "\n")[1:] {
		alan := strings.Fields(satir)
		if len(alan) > 3 && strings.HasSuffix(alan[1], arananPort) && alan[3] == "06" {
			say++
		}
	}
	return say
}

/*
DENEYİNİZ

 1. strace çıktısında Go'nun her net.Listen için yaptığı setsockopt çağrısını bulunuz.
 2. Programı arka arkaya iki kez çalıştırınız. İkinci çalıştırmada TIME_WAIT sayısı
    neden artmıştır?
 3. c.Close() ile sunucudaki c.Close() satırlarının sırasını değiştiriniz: istemci önce
    kapatsın. TIME_WAIT hangi tarafta kalmaktadır ve SO_REUSEADDR kapalıyken sunucu
    açılabilmekte midir?
 4. cat /proc/net/tcp çıktısını inceleyiniz. Adresler ve portlar hangi biçimde yazılmıştır?

SORULAR

  - TIME_WAIT neden vardır? Kaldırılsaydı hangi hata ortaya çıkabilirdi?
  - SO_REUSEADDR ile SO_REUSEPORT arasındaki fark nedir?
  - Ödevdeki çökme testlerinde (T4, T5) süreç kill -9 ile öldürüldükten hemen sonra
    yeniden başlatılmaktadır. C ile yazılmış bir sunucuda bu seçenek unutulsaydı ne olurdu?
*/
