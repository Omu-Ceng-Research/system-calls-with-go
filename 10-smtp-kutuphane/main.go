// Örnek 10 — SMTP kütüphanesiyle renkli ve ek dosyalı e-posta göndermek
//
// 2023-24 dönemindeki 4. hafta alıştırmasının ("Sending mails using SMTP protocol
// with Libraries") Go karşılığıdır:
//
//   - renkli kelimeler içeren bir e-posta göndermek,
//   - ekinde bir metin dosyası bulunan bir e-posta göndermek.
//
// İlgili sistem çağrıları: örnek 09 ile aynıdır; fark, protokolün kütüphanede kalmasıdır.
// Ödevde nerede: hazır bir kütüphanenin altında hangi soket konuşmasının yapıldığını görmek
//
// Çalıştırınız (Gmail hesabı gerekmeden, iki terminal):
//
//	ders 09 sahte-sunucu
//	ders 10 renkli -sunucu 127.0.0.1:2525 -alici ogrenci@ornek.com
//	ders 10 ekli   -sunucu 127.0.0.1:2525 -alici ogrenci@ornek.com -dosya go.mod
//
// Çalıştırınız (gerçek Gmail ile; GMAIL_ADRES ve GMAIL_SIFRE için örnek 09'a bakınız):
//
//	ders 10 renkli -alici arkadasiniz@ornek.com
//
// Örnek 09'da her SMTP satırı elle yazılmaktaydı. Burada net/smtp paketi aynı
// konuşmayı yapmaktadır: EHLO, gerekirse STARTTLS, AUTH, MAIL FROM, RCPT TO, DATA.
// Programın kendisi yalnızca e-postanın içeriğini (MIME biçiminde) hazırlamaktadır.
//
// Renkli metin için gövde HTML olarak gönderilmektedir. Ek dosya için gövde
// "multipart/mixed" olarak birden fazla parçaya bölünmekte ve dosya base64 ile
// kodlanmaktadır; çünkü SMTP yalnızca metin satırları taşımaktadır.
//
// Gözlemleyiniz (örnek 09'un sahte sunucusu açıkken):
//
//	cd /calisma
//	strace -f -e trace=network,read,write,close ./d10 renkli -sunucu 127.0.0.1:2525 -alici a@b.c
//
// Kütüphane kullanılması sistem çağrılarını değiştirmemektedir: strace çıktısı örnek
// 09'unkiyle neredeyse aynıdır. Yalnızca bu çağrıların hangi satırdan yapıldığı
// programın içinde değil, net/smtp paketinin içindedir.
package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Kullanım: d10 renkli|ekli -alici <adres> [-sunucu host:port] [-dosya yol]")
		os.Exit(1)
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	// Gmail'in 587 portu düz bağlantıyla açılıp STARTTLS ile şifrelenmektedir.
	// net/smtp.SendMail bunu kendiliğinden yapmaktadır.
	sunucu := fs.String("sunucu", "smtp.gmail.com:587", "SMTP sunucusu (host:port)")
	alici := fs.String("alici", "", "alıcının e-posta adresi")
	dosya := fs.String("dosya", "", "eklenecek metin dosyası (ekli için)")
	fs.Parse(os.Args[2:])
	if *alici == "" {
		log.Fatal("-alici verilmelidir")
	}

	gonderen := os.Getenv("GMAIL_ADRES")
	if gonderen == "" {
		gonderen = "gonderen@ornek.com"
	}

	var mesaj []byte
	switch os.Args[1] {
	case "renkli":
		mesaj = renkliEposta(gonderen, *alici)
	case "ekli":
		if *dosya == "" {
			log.Fatal("-dosya verilmelidir")
		}
		mesaj = ekliEposta(gonderen, *alici, *dosya)
	default:
		log.Fatal("bilinmeyen mod: ", os.Args[1])
	}

	// Kimlik doğrulama yalnızca şifre verildiyse yapılmaktadır. PlainAuth, şifreyi
	// şifrelenmemiş bir bağlantıda göndermeyi reddetmektedir (localhost hariç).
	var auth smtp.Auth
	if sifre := os.Getenv("GMAIL_SIFRE"); sifre != "" {
		host, _, _ := net.SplitHostPort(*sunucu)
		auth = smtp.PlainAuth("", gonderen, sifre, host)
	}
	// Bu tek satırın yaptığı sistem çağrıları (sahte sunucuyla):
	// → socket(AF_INET, SOCK_STREAM, ...) + connect(3, {"127.0.0.1", 2525}) ... (bkz. örnek 05)
	// → read(3, "220 sahte.sysprog ESMTP hazır\r\n", 4096)
	// → write(3, "EHLO localhost\r\n", 16)
	// → read(3, "250-sahte.sysprog merhaba\r\n", 4096)              çok satırlı cevap
	// → read(3, "250-8BITMIME\r\n250 AUTH LOGIN PLAIN\r\n", 4096)   iki read'de geldi
	// → write(3, "MAIL FROM:<gonderen@ornek.com> BODY=8BITMIME\r\n", 46)
	// → read(3, "250 Tamam\r\n"), write(3, "RCPT TO:<a@b.c>\r\n"), read(3, "250 Tamam\r\n")
	// → write(3, "DATA\r\n"), read(3, "354 ...")
	// → write(3, "From: ...", 398)                                   başlıklar, gövde ve "."
	// → read(3, "250 Kabul edildi\r\n"), write(3, "QUIT\r\n"), read(3, "221 ..."), close(3)
	// Sunucu STARTTLS sunsaydı (Gmail'in 587 portu gibi), EHLO'dan sonra şifreleme
	// başlatılır ve sonraki write'lardaki baytlar şifreli görünürdü.
	if err := smtp.SendMail(*sunucu, auth, gonderen, []string{*alici}, mesaj); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s e-posta gönderildi (%d bayt).\n", os.Args[1], len(mesaj))
}

func basliklar(b *bytes.Buffer, gonderen, alici, konu string) {
	fmt.Fprintf(b, "From: %s\r\n", gonderen)
	fmt.Fprintf(b, "To: %s\r\n", alici)
	// Türkçe karakter içeren konu satırı MIME ile kodlanmalıdır (=?UTF-8?q?...?=).
	fmt.Fprintf(b, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", konu))
	fmt.Fprintf(b, "MIME-Version: 1.0\r\n")
}

func renkliEposta(gonderen, alici string) []byte {
	var b bytes.Buffer
	basliklar(&b, gonderen, alici, "Renkli kelimeler — Sistem Programlama")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	b.WriteString(`<p>Bu e-postada <b style="color:#d32f2f">kırmızı</b>, ` +
		`<b style="color:#388e3c">yeşil</b> ve <b style="color:#1976d2">mavi</b> ` +
		`kelimeler bulunmaktadır.</p>` + "\r\n")
	b.WriteString(`<p style="color:#7b1fa2">Bu paragrafın tamamı mor renktedir.</p>` + "\r\n")
	return b.Bytes()
}

func ekliEposta(gonderen, alici, yol string) []byte {
	icerik, err := os.ReadFile(yol) // → openat + fstat + read + close (ek dosya)
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	basliklar(&b, gonderen, alici, "Ek dosyalı e-posta — Sistem Programlama")
	mw := multipart.NewWriter(&b)
	fmt.Fprintf(&b, "Content-Type: multipart/mixed; boundary=%s\r\n\r\n", mw.Boundary())

	// 1. parça: düz metin gövde
	metin, _ := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type": {"text/plain; charset=UTF-8"},
	})
	fmt.Fprintf(metin, "Merhaba,\r\n\r\nEkte %s dosyası bulunmaktadır.\r\n", filepath.Base(yol))

	// 2. parça: ek dosya, base64 ile 76 karakterlik satırlara bölünerek
	ek, _ := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/plain; charset=UTF-8"},
		"Content-Transfer-Encoding": {"base64"},
		"Content-Disposition":       {fmt.Sprintf("attachment; filename=%q", filepath.Base(yol))},
	})
	kodlu := base64.StdEncoding.EncodeToString(icerik)
	for len(kodlu) > 76 {
		fmt.Fprintf(ek, "%s\r\n", kodlu[:76])
		kodlu = kodlu[76:]
	}
	fmt.Fprintf(ek, "%s\r\n", kodlu)

	mw.Close() // kapanış sınırı: --<boundary>--
	return b.Bytes()
}

/*
DENEYİNİZ

 1. Sahte sunucunun ekranında net/smtp'nin gönderdiği komutları örnek 09'daki elle
    yazılmış komutlarla karşılaştırınız. Hangi komutlar farklıdır?
 2. ekli e-postada sahte sunucunun bastığı gövdede "boundary" satırlarını ve base64
    ile kodlanmış dosyayı bulunuz. Dosya base64 ile yüzde kaç büyümüştür?
 3. strace -f -c ./d10 renkli -sunucu 127.0.0.1:2525 -alici a@b.c ile kaç write(2)
    yapıldığını sayınız.

SORULAR

  - SMTP neden ikili (binary) veri taşıyamamaktadır? base64 bu sorunu nasıl çözmektedir?
  - Kütüphane kullanmanın kolaylığı açıktır. Örnek 09'daki gibi protokolü elle yazmak
    hangi durumda daha doğru bir seçim olur?
*/
