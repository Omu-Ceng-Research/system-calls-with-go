// 2023-24 dönemi örneği. Kodun kendisi değiştirilmemiştir; yalnızca yorumlar eklenmiştir.
//
// '→' ile başlayan yorumlar, altındaki satırın yaptığı sistem çağrılarını göstermektedir.
// Bunlar programın strace çıktısından alınmıştır (OpenJDK, Linux):
//
//     javac Client.java && strace -f -e trace=network,write,clone3 java Client
//
// JVM açılırken kendisi de yüzlerce sistem çağrısı yapmaktadır (sınıf dosyalarını okumak,
// GC ve JIT thread'lerini açmak, IPv6 ve SO_REUSEPORT desteğini yoklamak için socket +
// setsockopt + close). Bunlar programın satırlarına ait değildir; yukarıdaki strace komutu
// çıktısında 9991 portunun ilk geçtiği satırdan itibaren okunmalıdır.
//
// Go'dan farkı: Java soketleri bloklayan modda kullanmaktadır (Go'daki EAGAIN + epoll yoktur)
// ve her new Thread gerçek bir işletim sistemi thread'idir (clone3).

import java.io.*;
import java.net.*;
import java.util.Arrays;
import java.util.List;

public class Client {
    public static void main(String[] args) throws IOException {

        // "localhost" önce IP'ye çevrilmektedir (/etc/hosts okunmaktadır).
        // → socket(AF_INET6, SOCK_STREAM, IPPROTO_IP) = 5
        // → connect(5, {"::ffff:127.0.0.1", port 9991}) = 0     bloklayan connect
        // → getsockname(5)
        Socket socket = new Socket("localhost", 9991);
        // → write(1, "Connected!\n", 11)
        System.out.println("Connected!");

        Thread senderThread = new Thread(new Sender(socket));
        // → clone3(...)
        senderThread.start();

        List<Thread> customerList = Arrays.asList(new Thread[50]);
        // → 50 × clone3(...)   50 gerçek thread. strace -c ile programda toplam ~82 clone3
        // görülmektedir: 52'si bu programın, kalanı JVM'in kendi thread'leridir.
        customerList.forEach(d -> new Thread(new Sender(socket)).start() );

        System.out.println("thread sayisi: "+ customerList.size() );

        Thread receiverThread = new Thread(new Sender(socket));
        // → clone3(...)   adı receiver olsa da bu thread de yazmaktadır (Sender)
        receiverThread.start();
    }
}

class Sender implements Runnable {
    private Socket socket;

    public Sender(Socket socket) {
        this.socket = socket;
    }

    public void run() {
        while(true) {
            for (int i = 0; i < 10000000; i++) { }
            try {
                // true: autoflush. println her çağrıldığında tampon hemen sokete yazılmaktadır.
                PrintWriter out = new PrintWriter(socket.getOutputStream(), true);
                for (int i = 0; i < 10; i++) {
                    // → write(5, "Thread_id: 29 Message 2929...\n", 2023) = 2023
                    // Satırın tamamı tek write ile gitmektedir. 52 thread aynı fd 5'e yazmaktadır.
                    out.println("Thread_id: " + Thread.currentThread().getId() + " Message " + Long.toString(Thread.currentThread().getId()).repeat(1000));
                    // → write(1, "Message sended\n", 15)
                    System.out.println("Message sended");
                    //Thread.sleep(1000);
                }
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
    }
}
