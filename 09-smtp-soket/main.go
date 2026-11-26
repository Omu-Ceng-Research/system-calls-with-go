// Örnek 09 — SMTP ile e-posta göndermek: kütüphanesiz, doğrudan soket üzerinden
//
// 2023-24 dönemindeki "Sending mails using SMTP protocol via TELNET" örneğinin ve
// "A Small Exercise Project" alıştırmasının (kütüphane kullanmadan, telnet tabanlı
// bir e-posta istemcisi) Go karşılığıdır.
//
// İlgili sistem çağrıları: socket, connect, write, read (TLS bunların üzerinde çalışmaktadır)
// Ödevde nerede: metin tabanlı, satır satır ilerleyen bir protokol (3. bölüm) ve
// çok satırlı cevapların okunması (ASK cevabı)
//
// Çalıştırınız (Gmail hesabı gerekmeden, iki terminal):
//
//	ders 09 sahte-sunucu
//	ders 09 gonder -sunucu 127.0.0.1:2525 -tls=false -alici ogrenci@ornek.com
//
// Çalıştırınız (gerçek Gmail ile):
//
//	export GMAIL_ADRES=adiniz@gmail.com
//	export GMAIL_SIFRE='uygulama şifresi'        (aşağıdaki açıklamaya bakınız)
//	ders 09 gonder -alici arkadasiniz@ornek.com
//
// 2023-24'te aynı konuşma elle yapılmaktaydı:
//
//	openssl s_client -connect smtp.gmail.com:465 -crlf -ign_eof
//	EHLO localhost
//	AUTH LOGIN
//	<base64 ile kodlanmış adres>
//	<base64 ile kodlanmış uygulama şifresi>
//	MAIL FROM: <adres>
//	RCPT TO: <alıcı>
//	DATA
//	Subject: ...
//	<gövde>
//	.
//	QUIT
//
// Bu program aynı satırları sırayla göndermekte ve sunucunun her cevabını ekrana
// basmaktadır. Gönderilen satırlar "C:", gelen satırlar "S:" ile gösterilmektedir.
//
// Gözlemleyiniz (sahte sunucu açıkken):
//
//	cd /calisma
//	strace -f -e trace=network,read,write,close ./d09 gonder -sunucu 127.0.0.1:2525 -tls=false -alici a@b.c
//
// Kodun içindeki "→" ile başlayan yorumlar, o satırın yaptığı sistem çağrılarını
// göstermektedir ve strace çıktısından alınmıştır. Her SMTP komutu bir write(2),
// her cevap bir ya da birkaç read(2) olarak görünmektedir; strace çıktısı 2023-24'teki
// telnet oturumunun neredeyse aynısıdır.
//
// Gmail uygulama şifresi: myaccount.google.com → Güvenlik → 2 Adımlı Doğrulama
// açılmalıdır; ardından "Uygulama şifreleri" bölümünden yeni bir şifre üretilmektedir.
// Normal Gmail şifresi SMTP'de kabul edilmemektedir. Şifre koda yazılmamalı, ortam
// değişkeniyle verilmelidir.
package main

import (
	"bufio"           // Reader, NewReader
	"crypto/tls"      // Dial, Config
	"encoding/base64" // StdEncoding
	"flag"            // NewFlagSet, ExitOnError
	"fmt"             // Fprintln, Println, Print
	"log"             // Fatal, Fatalf
	"net"             // Conn, SplitHostPort, DialTimeout, Listen
	"os"              // Args, Stderr, Exit, Getenv
	"strings"         // TrimRight, HasPrefix, Split, ToUpper, SplitN, Contains, TrimPrefix
	"time"            // Second
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Kullanım: d09 gonder [bayraklar] | d09 sahte-sunucu [-port 2525]")
		os.Exit(1)
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	sunucu := fs.String("sunucu", "smtp.gmail.com:465", "SMTP sunucusu (host:port)")
	tlsKullan := fs.Bool("tls", true, "bağlantı baştan TLS ile mi açılsın (Gmail 465 için evet)")
	alici := fs.String("alici", "", "alıcının e-posta adresi")
	konu := fs.String("konu", "System Prog. - Go ile SMTP", "e-postanın konusu")
	mesaj := fs.String("mesaj", "Bu e-posta Go ile, doğrudan soket üzerinden gönderilmiştir.", "gövde")
	port := fs.String("port", "2525", "sahte sunucunun portu")
	fs.Parse(os.Args[2:])

	switch os.Args[1] {
	case "gonder":
		if *alici == "" {
			log.Fatal("-alici verilmelidir")
		}
		gonder(*sunucu, *tlsKullan, *alici, *konu, *mesaj)
	case "sahte-sunucu":
		sahteSunucu(*port)
	}
}

// ---------------------------------------------------------------------------
// İstemci

type oturum struct {
	r *bufio.Reader
	w net.Conn
}

// komut, bir satır göndermekte ve cevabı beklenen kodla karşılaştırmaktadır.
// SMTP'de satır sonu "\r\n"dir (openssl s_client'taki -crlf bayrağı bu yüzdendi).
func (o *oturum) komut(satir, gizli string, beklenen string) {
	if gizli != "" {
		fmt.Println("C:", gizli)
	} else {
		fmt.Println("C:", satir)
	}
	// → write(fd, "EHLO localhost\r\n", 16) = 16
	if _, err := o.w.Write([]byte(satir + "\r\n")); err != nil {
		log.Fatal(err)
	}
	o.cevapBekle(beklenen)
}

// cevapBekle, çok satırlı cevapları da okumaktadır. "250-..." ile başlayan satırlar
// devam satırıdır; "250 ..." (koddan sonra boşluk) son satırdır.
func (o *oturum) cevapBekle(beklenen string) {
	for {
		// bufio.Reader, 4096 baytlık tamponu boşsa read(2) yapmaktadır; doluysa
		// satırı tampondan vermektedir. Çok satırlı "250-..." cevabı bazen tek read'de
		// gelmektedir:
		// → read(fd, "250-sahte.sysprog merhaba\r\n250-8BITMIME\r\n250 AUTH ...", 4096) = 63
		// bazen de iki read'de (örnek 10'un strace çıktısında görüldüğü gibi):
		// → read(fd, "250-sahte.sysprog merhaba\r\n", 4096) = 27
		// → read(fd, "250-8BITMIME\r\n250 AUTH LOGIN PLAIN\r\n", 4096) = 36
		// Satır satır okumak bu farkı ortadan kaldırmaktadır (bkz. örnek 05, SORULAR).
		satir, err := o.r.ReadString('\n')
		if err != nil {
			log.Fatal("sunucu bağlantıyı kapattı: ", err)
		}
		satir = strings.TrimRight(satir, "\r\n")
		fmt.Println("S:", satir)
		if len(satir) < 4 {
			log.Fatal("geçersiz cevap")
		}
		if satir[3] == ' ' {
			if !strings.HasPrefix(satir, beklenen) {
				log.Fatalf("beklenen %s, gelen %s", beklenen, satir[:3])
			}
			return
		}
	}
}

func gonder(sunucu string, tlsKullan bool, alici, konu, mesaj string) {
	adres := os.Getenv("GMAIL_ADRES")
	sifre := os.Getenv("GMAIL_SIFRE")
	if adres == "" {
		adres = "gonderen@ornek.com"
	}

	var conn net.Conn
	var err error
	if tlsKullan {
		// tls.Dial: önce TCP bağlantısı (socket + connect, bkz. örnek 05), ardından
		// TLS el sıkışması (birkaç write ve read). openssl s_client'ın yaptığı iş budur.
		// TLS çekirdekte değil, programın içinde (Go'nun crypto/tls paketinde)
		// yapılmaktadır. Bu yüzden strace'te write(2) ile giden baytlar şifrelidir;
		// "EHLO localhost" düz metin olarak görünmemektedir.
		host, _, _ := net.SplitHostPort(sunucu)
		conn, err = tls.Dial("tcp", sunucu, &tls.Config{ServerName: host})
	} else {
		// → socket(AF_INET, SOCK_STREAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_IP) = 5
		// → connect(5, {"127.0.0.1", port 2525}) = -1 EINPROGRESS ... (bkz. örnek 05)
		// Zaman aşımı sistem çağrısıyla değil, epoll_pwait'in bekleme süresiyle uygulanmaktadır.
		conn, err = net.DialTimeout("tcp", sunucu, 5*time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close() // → close(5)

	o := &oturum{r: bufio.NewReader(conn), w: conn}
	o.cevapBekle("220") // sunucu önce kendini tanıtmaktadır

	o.komut("EHLO localhost", "", "250")
	if sifre != "" {
		b64 := base64.StdEncoding.EncodeToString
		o.komut("AUTH LOGIN", "", "334")
		o.komut(b64([]byte(adres)), "<adres, base64>", "334")
		o.komut(b64([]byte(sifre)), "<şifre, base64 - ekrana basılmadı>", "235")
	}
	o.komut("MAIL FROM:<"+adres+">", "", "250")
	o.komut("RCPT TO:<"+alici+">", "", "250")
	o.komut("DATA", "", "354")

	// Başlıklar ve gövde. Türkçe karakterler için içerik türü belirtilmektedir.
	govde := "From: " + adres + "\r\n" +
		"To: " + alici + "\r\n" +
		"Subject: " + konu + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n"
	for _, satir := range strings.Split(mesaj, "\n") {
		// Tek başına "." gövdenin sonu demektir. Nokta ile başlayan satırın başına
		// bir nokta daha eklenmektedir (dot-stuffing).
		if strings.HasPrefix(satir, ".") {
			satir = "." + satir
		}
		govde += satir + "\r\n"
	}
	fmt.Print("C: <başlıklar ve gövde>\n")
	conn.Write([]byte(govde)) // → write(fd, "From: ...\r\nTo: ...\r\n...", 201): başlıklar ve gövde tek write
	o.komut(".", "", "250")
	o.komut("QUIT", "", "221")
	fmt.Println("E-posta gönderildi.")
}

// ---------------------------------------------------------------------------
// Sahte sunucu: e-postayı göndermemekte, yalnızca gelen konuşmayı ekrana basmaktadır.
// Örnek 10'daki kütüphaneli istemci de bununla denenebilmektedir.

func sahteSunucu(port string) {
	// Adres verildiği için IPv6 değil, doğrudan IPv4 soketi açılmaktadır:
	// → socket(AF_INET, SOCK_STREAM, ...) + setsockopt(SO_REUSEADDR) + bind({"127.0.0.1", 2525}) + listen
	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sahte SMTP sunucusu dinliyor:", ln.Addr())
	for {
		conn, err := ln.Accept() // → accept4(3, ...) = 4
		if err != nil {
			continue
		}
		go sahteOturum(conn)
	}
}

func sahteOturum(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	yaz := func(s string) { conn.Write([]byte(s + "\r\n")) } // → her cevap satırı bir write(4, ...)
	oku := func() (string, bool) {
		s, err := r.ReadString('\n')
		return strings.TrimRight(s, "\r\n"), err == nil
	}

	fmt.Println("---", conn.RemoteAddr(), "bağlandı")
	yaz("220 sahte.sysprog ESMTP hazır")
	for {
		satir, ok := oku()
		if !ok {
			return
		}
		fmt.Println("C:", satir)
		komut := strings.ToUpper(strings.SplitN(satir, " ", 2)[0])
		switch {
		case komut == "EHLO":
			yaz("250-sahte.sysprog merhaba")
			yaz("250-8BITMIME")
			yaz("250 AUTH LOGIN PLAIN")
		case komut == "HELO":
			yaz("250 sahte.sysprog merhaba")
		case komut == "AUTH" && strings.Contains(strings.ToUpper(satir), "LOGIN"):
			yaz("334 VXNlcm5hbWU6") // "Username:" base64
			oku()
			yaz("334 UGFzc3dvcmQ6") // "Password:" base64
			oku()
			yaz("235 Kimlik doğrulandı (sahte sunucu her şeyi kabul etmektedir)")
		case komut == "AUTH":
			yaz("235 Kimlik doğrulandı")
		case komut == "MAIL", komut == "RCPT", komut == "RSET", komut == "NOOP":
			yaz("250 Tamam")
		case komut == "DATA":
			yaz("354 Gövdeyi gönderiniz, tek başına '.' ile bitiriniz")
			fmt.Println("----- e-posta -----")
			for {
				s, ok := oku()
				if !ok {
					return
				}
				if s == "." {
					break
				}
				fmt.Println("  ", strings.TrimPrefix(s, "."))
			}
			fmt.Println("-------------------")
			yaz("250 Kabul edildi")
		case komut == "QUIT":
			yaz("221 Güle güle")
			return
		default:
			yaz("502 Komut bilinmiyor")
		}
	}
}

/*
DENEYİNİZ

 1. Sahte sunucuya telnet yerine nc ile elle bağlanınız ve 2023-24'teki adımları
    kendiniz yazınız:  nc -C 127.0.0.1 2525     (-C: satır sonu \r\n)
 2. strace -f -e trace=connect,write,read ./d09 gonder -sunucu 127.0.0.1:2525 -tls=false -alici a@b.c
    çıktısında her SMTP komutunun ayrı bir write(2) olduğunu doğrulayınız.
 3. Gerçek Gmail ile gönderirken -tls=false deneyiniz. Ne olmaktadır?
 4. "-mesaj" içine tek başına "." olan bir satır koyunuz:
    -mesaj "$(printf 'satır 1\n.\nsatır 3')"
    Dot-stuffing kaldırılsaydı sunucu e-postayı nerede bitmiş sayardı?

SORULAR

  - base64 bir şifreleme midir? AUTH LOGIN'in güvenliğini sağlayan katman hangisidir?
  - cevapBekle neden tek bir satır değil, kodun ardından boşluk gelen satıra kadar
    okumaktadır? Ödevdeki ASK cevabı da çok satırlıdır; istemci kaç satır okuyacağını
    nereden bilmektedir?
*/
