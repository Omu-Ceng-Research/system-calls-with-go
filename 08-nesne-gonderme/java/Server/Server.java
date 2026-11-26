// 2023-24 dönemi örneği. Kodun kendisi değiştirilmemiştir; yalnızca yorumlar eklenmiştir.
//
// '→' ile başlayan yorumlar, altındaki satırın yaptığı sistem çağrılarını göstermektedir.
// Bunlar programın strace çıktısından alınmıştır (OpenJDK, Linux):
//
//     javac *.java && strace -f -e trace=network,read,write java Server
//
// JVM açılırken kendisi de yüzlerce sistem çağrısı yapmaktadır (sınıf dosyalarını okumak,
// GC ve JIT thread'lerini açmak, IPv6 ve SO_REUSEPORT desteğini yoklamak için socket +
// setsockopt + close). Bunlar programın satırlarına ait değildir; yukarıdaki strace komutu
// çıktısında 9991 portunun ilk geçtiği satırdan itibaren okunmalıdır.
//

import java.io.*;
import java.net.*;
import java.util.concurrent.locks.ReentrantLock;

public class Server {
    public static void main(String[] args)  {

        try {
            // → socket(AF_INET6, SOCK_STREAM, IPPROTO_IP) = 5
            // → setsockopt(5, IPV6_V6ONLY, [0]), setsockopt(5, SO_REUSEADDR, [1])
            // → bind(5, {"::", port 9991}), listen(5, 50)
            ServerSocket serverSocket = new ServerSocket(9991);
            // → write(1, "Merhaba\n", 8)
            System.out.println("Merhaba");
            while (true) {

                // → accept(5, ...) = 6
                Socket clientSocket = serverSocket.accept();

                InputStream inputStream = clientSocket.getInputStream();
                // Yapıcı, akışın başlığını hemen okumaktadır:
                // → read(6, "\254\355", 2)    sihirli sayı 0xACED
                // → read(6, "\0\5", 2)        sürüm 5
                ObjectInputStream objectInputStream = new ObjectInputStream(inputStream);
                // → read(6, "s", 1), read(6, "r", 1), read(6, "\0\6", 2), read(6, "Person", 6),
                //   read(6, "\0\0\0\0\0\0\0\1", 8) ...   (serialVersionUID = 1)
                // Akış tamponsuz olduğu için ObjectInputStream her alanı ayrı bir read ile okumaktadır:
                // 86 baytlık tek nesne için toplam 24 read yapılmaktadır. new BufferedInputStream(inputStream)
                // ile sarılsaydı tek read yeterdi (Go sürümünde bufio.Reader bunu yapmaktadır).
                Person incomingPerson = (Person) objectInputStream.readObject();

                // → write(1, ...)   her println bir write
                System.out.println("Gelen kisi nesnesinin adi: " + incomingPerson.getName());
                // height transient olduğu için gönderilmemiştir; 0 basılmaktadır.
                System.out.println("Gelen kisi nesnesinin boyu: "+incomingPerson.getHeight());
                System.out.println("Gelen kisi nesnesinin yasi: "+incomingPerson.getAge());

                // → close(6)
                clientSocket.close();
                // → close(5)   Dinleme soketi kapatılmaktadır. Döngünün bir sonraki accept'i
                // "java.net.SocketException: Socket is closed" hatası vermekte ve program bitmektedir.
                serverSocket.close();
            }
        } catch (IOException e) {
            e.printStackTrace();
        } catch (ClassNotFoundException e) {
            throw new RuntimeException(e);
        }
    }
}