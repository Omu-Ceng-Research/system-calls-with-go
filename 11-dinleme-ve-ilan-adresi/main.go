// Örnek 11 — Dinleme adresi ve ilan adresi: 0.0.0.0, 127.0.0.1 ve gerçek IP
//
// İlgili sistem çağrıları: bind, listen, connect, getsockname
// Ödevde nerede: 5.4 (-dinle 0.0.0.0:7100 ile -ilan 10.0.0.12:7100 ayrımı), T8 laboratuvar demosu
//
// Çalıştırınız:
//
//	ders 11                            (varsayılan: -dinle 127.0.0.1:7100)
//	ders 11 -dinle 0.0.0.0:7100
//	ders 11 -dinle <bu makinenin IP'si>:7100
//
// Gözlemleyiniz:
//
//	cd /calisma
//	strace -f -e trace=network,close ./d11
//
// Kodun içindeki "→" ile başlayan yorumlar, o satırın yaptığı sistem çağrılarını
// göstermektedir ve strace çıktısından alınmıştır.
//
// bind(2) bir soketi bir adrese bağlamaktadır. Hangi adrese bağlandığı, sokete
// kimlerin bağlanabileceğini belirlemektedir:
//
//	127.0.0.1:7100   yalnızca aynı makineden (loopback arayüzü)
//	10.0.0.12:7100   yalnızca bu IP'ye gelen bağlantılar (ve aynı makineden bu IP ile)
//	0.0.0.0:7100     makinenin bütün arayüzlerinden
//
// 0.0.0.0 bir dinleme adresidir, bağlanılacak bir adres değildir. Bir düğüm
// koordinatöre "bana 0.0.0.0:7100 adresinden ulaşınız" diyemez. Bu yüzden ödevde
// iki ayrı adres vardır: düğüm 0.0.0.0'ı dinlemekte, başkalarına ise kendi gerçek
// IP'sini ilan etmektedir.
//
// Docker'da durum bir adım daha karmaşıktır. Konteynerin içindeki IP (ör. 172.17.0.2)
// yalnızca o bilgisayarın içinden görünmektedir. Laboratuvardaki başka bir makine
// konteynere, ancak port yönlendirmesi (docker run -p 7100:7100) ile ve konteynerin
// değil bilgisayarın IP'si üzerinden ulaşabilmektedir. Bu yüzden konteynerdeki düğüm
// -ilan olarak bilgisayarın laboratuvar IP'sini vermelidir.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	dinle := flag.String("dinle", "127.0.0.1:7100", "dinlenecek adres")
	flag.Parse()

	fmt.Println("Bu makinenin ağ arayüzleri:")
	var adresler []string
	// Arayüz listesi bir dosyadan değil, çekirdekle konuşan özel bir soketten
	// (netlink) alınmaktadır. ip addr komutu da aynı yolu kullanmaktadır:
	// → socket(AF_NETLINK, SOCK_RAW|SOCK_CLOEXEC, NETLINK_ROUTE) = 3
	// → bind(3, {AF_NETLINK}), getsockname(3)
	// → sendto(3, [{nlmsg_type=RTM_GETLINK, ...}])        "arayüzleri listele" isteği
	// → recvfrom(3, [{nlmsg_type=RTM_NEWLINK, ...}, ...]) her arayüz için bir mesaj
	// → recvfrom(3, [{nlmsg_type=NLMSG_DONE}]), close(3)
	arayuzler, _ := net.Interfaces()
	for _, a := range arayuzler {
		ipler, _ := a.Addrs() // → aynı yol, RTM_GETADDR isteğiyle (her çağrıda yeni bir netlink soketi)
		for _, ip := range ipler {
			if ipn, ok := ip.(*net.IPNet); ok && ipn.IP.To4() != nil {
				fmt.Printf("  %-10s %s\n", a.Name, ipn.IP)
				adresler = append(adresler, ipn.IP.String())
			}
		}
	}
	fmt.Println("\nİlan için önerilen adres (dışarıya çıkan arayüz):", disariCikanIP())
	if _, err := os.Stat("/.dockerenv"); err == nil {
		fmt.Println("UYARI: Program Docker içinde çalışmaktadır. Bu adres yalnızca bu bilgisayarın")
		fmt.Println("içinden görünmektedir; laboratuvarda bilgisayarınızın kendi IP'si ilan edilmelidir.")
	}

	// Adres verildiği için doğrudan IPv4 soketi açılmaktadır (":7100" verilseydi
	// IPv6 çift yığın soket açılırdı, bkz. örnek 05):
	// → socket(AF_INET, SOCK_STREAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_IP) = 3
	// → setsockopt(3, SO_REUSEADDR), bind(3, {"127.0.0.1", port 7100}), listen(3, 4096)
	// → getsockname(3)
	// bind'a verilen adres, hangi arayüzden gelen bağlantıların kabul edileceğini
	// belirlemektedir. Aşağıdaki deneme bunu göstermektedir.
	ln, err := net.Listen("tcp", *dinle)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	fmt.Printf("\nDinlenen adres: %s\n\n", ln.Addr())

	fmt.Println("Her arayüzün IP'si üzerinden bağlanmayı deniyorum:")
	for _, ip := range adresler {
		hedef := net.JoinHostPort(ip, port)
		// → socket(AF_INET, ...) = 6
		// → connect(6, {"172.17.0.2", port 7100}) = -1 EINPROGRESS
		// → getsockopt(6, SO_ERROR) = ECONNREFUSED   o IP'de dinleyen yoksa
		c, err := net.DialTimeout("tcp", hedef, time.Second)
		if err != nil {
			fmt.Printf("  %-21s BAĞLANILAMADI  (%v)\n", hedef, err)
			continue
		}
		c.Close()
		fmt.Printf("  %-21s bağlanıldı\n", hedef)
	}
}

// disariCikanIP, varsayılan rotanın hangi yerel IP'den geçtiğini bulmaktadır.
// UDP'de connect(2) hiçbir paket göndermemektedir; çekirdek yalnızca rotayı seçmekte
// ve soketin yerel adresini doldurmaktadır. Bu adres getsockname(2) ile okunmaktadır.
func disariCikanIP() string {
	// → socket(AF_INET, SOCK_DGRAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_IP) = 3
	// → connect(3, {"8.8.8.8", port 80}) = 0      ağa paket gitmemektedir (bkz. örnek 06)
	// → getsockname(3) = {"172.17.0.2", port 42777}
	c, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "(bulunamadı: " + err.Error() + ")"
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).IP.String()
}

/*
DENEYİNİZ

 1. Üç dinleme adresini sırayla deneyiniz ve hangi IP'lerden bağlanılabildiğini
    karşılaştırınız.
 2. Aynı port için iki programı aynı anda çalıştırınız: biri 127.0.0.1:7100, diğeri
    0.0.0.0:7100 dinlesin. İkincisi başlayabilmekte midir? Hata mesajı nedir?
 3. Konteynerin dışında (kendi bilgisayarınızda) ipconfig (Windows) ya da
    ip addr (Linux) / ifconfig (macOS) ile IP'nizi bulunuz. Konteynerin içindeki IP
    ile aynı mıdır?
 4. Laboratuvarda iki makine ile: birinde ders 11 -dinle 0.0.0.0:7100 çalışırken
    diğerinden nc -v <ilk makinenin IP'si> 7100 ile bağlanınız.

SORULAR

  - Bir düğüm -ilan adresi olarak 127.0.0.1 verirse koordinatör aynı makinedeyken
    çalışmakta, laboratuvarda ise çalışmamaktadır. Neden?
  - Docker Desktop'ta konteyner içindeki IP'ye dışarıdan neden ulaşılamamaktadır?
*/
