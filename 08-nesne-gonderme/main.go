// Örnek 08 — Soket üzerinden nesne göndermek: gob, JSON ve elle ikili biçim
//
// 2023-24 dönemindeki "Sending objects over TCP Sockets with Java" örneğinin
// (Person.java, Server.java, Client.java) Go karşılığıdır.
//
// İlgili sistem çağrıları: connect, write, read
// Ödevde nerede: 5.1 (metin.dat ve vektor.dat için elle ikili biçim), gRPC/protobuf mesajları
//
// Çalıştırınız (iki terminal):
//
//	ders 08 sunucu
//	ders 08 istemci -bicim gob        (Java'daki ObjectOutputStream'in karşılığı)
//	ders 08 istemci -bicim json
//	ders 08 istemci -bicim ikili      (alan alan, elle: ödevdeki dosya biçimleri gibi)
//
// Gözlemleyiniz:
//
//	cd /calisma
//	strace -f -e trace=read ./d08 sunucu
//	strace -f -e trace=write ./d08 istemci -bicim gob
//
// Kodun içindeki "→" ile başlayan yorumlar, o satırın yaptığı sistem çağrılarını
// göstermektedir ve strace çıktısından alınmıştır. Bağlantı kurma ve kapatma
// çağrıları örnek 05'teki ile aynıdır.
//
// Java'da bir nesnenin sokete yazılabilmesi için sınıfın Serializable olması
// gerekmekteydi. Soketin kendisi nesne tanımamaktadır; yalnızca bayt taşımaktadır.
// Nesneyi bayta çevirme işine serileştirme (serialization) denmektedir ve bunun
// birden fazla yolu vardır:
//
//	Biçim   Karşılığı                     Başka dil okuyabilir mi?   Boyut
//	gob     Java Serializable             Hayır (yalnızca Go)         orta
//	json    —                             Evet                        büyük
//	ikili   protobuf'un elle yapılanı     Evet, biçim belgelenirse    en küçük
//
// Java'daki iki anahtar kelimenin Go karşılığı:
//
//   - transient int height: Go'da küçük harfle başlayan (dışa aktarılmayan) alanlar
//     gob ve JSON tarafından serileştirilmemektedir. Burada boy alanı bu yüzden küçük
//     harfle başlamaktadır ve karşı tarafa 0 olarak ulaşmaktadır.
//   - static String country: Go'da paket düzeyindeki bir değişkendir; nesnenin
//     parçası olmadığı için hiçbir biçimde gönderilmemektedir.
package main

import (
	"bufio"           // NewReader
	"encoding/binary" // LittleEndian, Read
	"encoding/gob"    // NewEncoder, NewDecoder
	"encoding/json"   // NewEncoder, NewDecoder
	"flag"            // NewFlagSet, ExitOnError
	"fmt"             // Fprintln, Println, Printf, Errorf
	"io"              // Writer, Reader, ReadFull
	"log"             // Fatal, Print
	"net"             // Dial, JoinHostPort, Listen, Conn
	"os"              // Args, Stderr, Exit
)

const port = "9991"

// Ulke, Java'daki static country alanının karşılığıdır.
var Ulke = "TR"

type Kisi struct {
	Ad  string
	Yas int
	boy int // Java'daki transient: serileştirilmemektedir
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Kullanım: d08 sunucu | d08 istemci [-bicim gob|json|ikili]")
		os.Exit(1)
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	bicim := fs.String("bicim", "gob", "serileştirme biçimi: gob, json, ikili")
	adres := fs.String("adres", "127.0.0.1", "sunucunun adresi")
	fs.Parse(os.Args[2:])

	switch os.Args[1] {
	case "sunucu":
		sunucu()
	case "istemci":
		istemci(*adres, *bicim)
	}
}

func istemci(adres, bicim string) {
	// → socket + connect + ... (bkz. örnek 05)
	conn, err := net.Dial("tcp", net.JoinHostPort(adres, port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close() // → close(3)
	fmt.Println("Bağlanıldı!")

	kisi := Kisi{Ad: "ismail hakkı turan", Yas: 32, boy: 190}

	// Gönderilen baytları saymak için bağlantının önüne bir sayaç konmaktadır.
	sayac := &sayanYazici{w: conn}
	// İlk bayt sunucuya biçimi bildirmektedir: g, j ya da i.
	sayac.Write([]byte{bicim[0]}) // → write(3, "g", 1) = 1

	// Serileştirme bellekte yapılmaktadır; sokete yazılan, sonuçta oluşan baytlardır.
	// Kütüphanenin kaç write(2) yaptığı biçime göre değişmektedir:
	switch bicim {
	case "gob":
		// → write(3, " \177\3\1\1\4Kisi\1...Ad...Yas...", 33)   önce tür tanımı (alan adları)
		// → write(3, "\32\377\200\1\23ismail hakkı turan\1@\0", 27)   sonra değerler
		err = gob.NewEncoder(sayac).Encode(kisi)
	case "json":
		// → write(3, "{\"Ad\":\"ismail hakkı turan\",\"Yas\":32}\n", 38)
		err = json.NewEncoder(sayac).Encode(kisi)
	case "ikili":
		// → write(3, "\23\0\0\0ismail hakkı turan \0\0\0", 27)
		err = ikiliYaz(sayac, kisi)
	default:
		log.Fatal("bilinmeyen biçim: ", bicim)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Nesne gönderildi (%s, %d bayt). Soket kapatılıyor.\n", bicim, sayac.n)
}

func sunucu() {
	// → socket + setsockopt × 2 + bind + listen (bkz. örnek 05)
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Merhaba, port", port)
	// Java örneğinde sunucu ilk nesneden sonra serverSocket.close() ile kapanmaktaydı.
	// Burada her bağlantı sırayla kabul edilmektedir.
	for {
		conn, err := ln.Accept() // → accept4(3, ...) = 4 + setsockopt × 5
		if err != nil {
			log.Print(err)
			continue
		}
		nesneAl(conn)
	}
}

func nesneAl(conn net.Conn) {
	defer conn.Close() // → close(4)
	r := bufio.NewReader(conn)
	// bufio.Reader ilk okumada 4096 baytlık tamponunu doldurmaya çalışmaktadır.
	// Nesne küçük olduğu için biçim baytı ve nesnenin tamamı tek read(2) ile gelmektedir:
	// → read(4, "g \177\3\1\1\4Kisi...", 4096) = 61
	// Aşağıdaki ReadByte ve Decode çağrıları bu tampondan okumakta, sistem çağrısı
	// yapmamaktadır. Nesne 4096 bayttan büyük olsaydı ek read'ler görülürdü.
	tur, err := r.ReadByte()
	if err != nil {
		log.Print(err)
		return
	}
	var kisi Kisi
	switch tur {
	case 'g':
		err = gob.NewDecoder(r).Decode(&kisi)
	case 'j':
		err = json.NewDecoder(r).Decode(&kisi)
	case 'i':
		kisi, err = ikiliOku(r)
	default:
		err = fmt.Errorf("bilinmeyen biçim baytı %q", tur)
	}
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println("Gelen kişi nesnesinin adı:", kisi.Ad)
	fmt.Println("Gelen kişi nesnesinin boyu:", kisi.boy, "(transient: gönderilmedi)")
	fmt.Println("Gelen kişi nesnesinin yaşı:", kisi.Yas)
	fmt.Println()
}

// ikiliYaz, nesneyi alan alan, sabit bir düzenle yazmaktadır:
//
//	[ad uzunluğu: uint32][ad: UTF-8 baytları][yaş: uint32]    (little-endian)
//
// Ödevdeki metin.dat kayıtları da aynı yöntemle yazılmaktadır.
func ikiliYaz(w io.Writer, k Kisi) error {
	ad := []byte(k.Ad)
	tampon := make([]byte, 4+len(ad)+4)
	binary.LittleEndian.PutUint32(tampon[0:], uint32(len(ad)))
	copy(tampon[4:], ad)
	binary.LittleEndian.PutUint32(tampon[4+len(ad):], uint32(k.Yas))
	_, err := w.Write(tampon) // → tek write(2): başlık, ad ve yaş birlikte
	return err
}

func ikiliOku(r io.Reader) (Kisi, error) {
	var uzunluk uint32
	if err := binary.Read(r, binary.LittleEndian, &uzunluk); err != nil {
		return Kisi{}, err
	}
	if uzunluk > 4096 { // karşı taraftan gelen uzunluğa körü körüne güvenilmemelidir
		return Kisi{}, fmt.Errorf("ad çok uzun: %d bayt", uzunluk)
	}
	ad := make([]byte, uzunluk)
	// io.ReadFull, istenen bayt sayısı tamamlanana kadar read'i tekrarlamaktadır.
	// Burada bufio tamponundan okunduğu için ek bir read(2) gerekmemektedir.
	if _, err := io.ReadFull(r, ad); err != nil {
		return Kisi{}, err
	}
	var yas uint32
	if err := binary.Read(r, binary.LittleEndian, &yas); err != nil {
		return Kisi{}, err
	}
	return Kisi{Ad: string(ad), Yas: int(yas)}, nil
}

type sayanYazici struct {
	w io.Writer
	n int
}

func (s *sayanYazici) Write(p []byte) (int, error) {
	n, err := s.w.Write(p)
	s.n += n
	return n, err
}

/*
DENEYİNİZ

 1. Üç biçimin gönderdiği bayt sayılarını karşılaştırınız.
 2. Sunucu yerine nc ile dinleyerek gönderilen baytları inceleyiniz:
    nc -l -p 9991 | hexdump -C         (ikinci terminalde: ders 08 istemci -bicim json)
    Aynısını gob ve ikili için tekrarlayınız. Hangisinde alan adları ("Ad", "Yas")
    baytların içinde yer almaktadır?
 3. Kisi yapısına yeni bir alan (ör. Sehir string) ekleyiniz ve yalnızca istemciyi
    yeniden derleyiniz. Eski sunucu hangi biçimlerde hata vermektedir?
 4. ikiliOku içindeki uzunluk denetimini kaldırınız ve şu komutla sunucuya
    4 GB'lık bir ad uzunluğu gönderiniz:
    printf 'i\xff\xff\xff\xff' | nc -q1 127.0.0.1 9991

SORULAR

  - Java'daki serialVersionUID neyi çözmeye çalışmaktadır? JSON ve gob'da bu sorun
    nasıl ele alınmaktadır?
  - Ödevde koordinatör ile düğüm arasında protobuf kullanılmaktadır. protobuf bu üç
    biçimden hangisine en çok benzemektedir?
*/
