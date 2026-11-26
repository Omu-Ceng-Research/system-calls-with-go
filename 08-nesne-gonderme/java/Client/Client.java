// 2023-24 dönemi örneği. Kodun kendisi değiştirilmemiştir; yalnızca yorumlar eklenmiştir.
//
// '→' ile başlayan yorumlar, altındaki satırın yaptığı sistem çağrılarını göstermektedir.
// Bunlar programın strace çıktısından alınmıştır (OpenJDK, Linux):
//
//     javac *.java && strace -f -e trace=network,write java Client
//
// JVM açılırken kendisi de yüzlerce sistem çağrısı yapmaktadır (sınıf dosyalarını okumak,
// GC ve JIT thread'lerini açmak, IPv6 ve SO_REUSEPORT desteğini yoklamak için socket +
// setsockopt + close). Bunlar programın satırlarına ait değildir; yukarıdaki strace komutu
// çıktısında 9991 portunun ilk geçtiği satırdan itibaren okunmalıdır.
//

import java.io.*;
import java.net.*;
import java.util.Arrays;
import java.util.List;

public class Client {
    public static void main(String[] args) throws IOException {

        // → socket(AF_INET6, SOCK_STREAM, IPPROTO_IP) = 5
        // → connect(5, {"::ffff:127.0.0.1", port 9991}) = 0
        Socket socket = new Socket("localhost", 9991);
        System.out.println("Connected!");

        OutputStream outputStream = socket.getOutputStream();
        // Yapıcı, akışın başlığını hemen yazmaktadır:
        // → write(5, "\254\355\0\5", 4)    0xACED (sihirli sayı) + sürüm 5
        ObjectOutputStream objectOutputStream = new ObjectOutputStream(outputStream);

        Person person = new Person();
        person.setHeight(190);
        person.setName("ismail hakki turan");
        person.setAge(32);

        System.out.println("Sending an object over the Socket");
        // → write(5, "sr\0\6Person\0\0\0\0\0\0\0\1...I\0\3age...", 55)   sınıf tanımı
        // → write(5, "xp", 2)
        // → write(5, "\0\0\0 ", 4)                                   age = 32
        // → write(5, "t\0\22ismail hakki turan", 21)                  name
        // height transient olduğu için gönderilmemektedir; static country de gönderilmemektedir.
        // Akış tamponsuz olduğu için bir nesne dört ayrı write ile gitmektedir.
        objectOutputStream.writeObject(person);

        System.out.println("Closing socket and terminating program.");
        // → shutdown(5, SHUT_WR) + close(5)
        socket.close();
    }
}