// 2023-24 dönemi örneği. Kodun kendisi değiştirilmemiştir; yalnızca yorumlar eklenmiştir.
//
// '→' ile başlayan yorumlar, altındaki satırın yaptığı sistem çağrılarını göstermektedir.
// Bunlar programın strace çıktısından alınmıştır (OpenJDK, Linux):
//
//     javac Server.java && strace -f -e trace=network,read,write,clone3 java Server
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
import java.util.concurrent.locks.ReentrantLock;

public class Server {
    public static void main(String[] args)  {

        try {
            // Create a server socket
            // → socket(AF_INET6, SOCK_STREAM, IPPROTO_IP) = 5
            // → setsockopt(5, SOL_IPV6, IPV6_V6ONLY, [0])      IPv4 bağlantıları da kabul edilsin
            // → setsockopt(5, SOL_SOCKET, SO_REUSEADDR, [1])
            // → bind(5, {"::", port 9991}), listen(5, 50)      Java'nın varsayılan kuyruğu 50
            ServerSocket serverSocket = new ServerSocket(9991);
            // → write(1, "Merhaba\n", 8)
            System.out.println("Merhaba");
            while (true) {

                // → accept(5, {"::ffff:127.0.0.1", port 54094}, ...) = 6
                // → getsockname(6)
                // Bağlanan olana kadar bu thread çekirdekte uyumaktadır. Her istemci için yeni bir fd.
                Socket clientSocket = serverSocket.accept();

                // Nesne oluşturulmaktadır; sistem çağrısı yoktur.
                Thread serverThread = new Thread(new ClientHandler(clientSocket));
                // → clone3(...)   yeni bir işletim sistemi thread'i (Go'daki go ifadesi clone yapmamaktadır)
                serverThread.start();

                // Bu satır ikinci, kullanılmayan bir ClientHandler nesnesi oluşturmaktadır; etkisi yoktur.
                new ClientHandler(clientSocket);
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}

class ClientHandler implements Runnable {
    private Socket clientSocket;

    public ClientHandler(Socket socket) {
        this.clientSocket = socket;
    }

    // Her ClientHandler'ın KENDİ kilidi vardır; kilidi kimse paylaşmadığı için hiçbir şeyi
    // korumamaktadır (bkz. ../main.go). Kilit boştayken lock/unlock sistem çağrısı yapmaz;
    // çekişme olsaydı bekleyen thread futex(FUTEX_WAIT) ile uyutulurdu.
    ReentrantLock lock = new ReentrantLock();

    public void run() {
        try {
            // Get the input and output streams for the socket
            // Akışlar (stream) yalnızca nesnedir; okuma yapılana kadar sistem çağrısı yoktur.
            BufferedReader in = new BufferedReader(new InputStreamReader(clientSocket.getInputStream()));
            PrintWriter out = new PrintWriter(clientSocket.getOutputStream(), true);

            // Communicate with the client
            while (true) {
                lock.lock();
                // → read(6, "Thread_id: 29 Message 2929...", 8192) = 2023
                // BufferedReader her satır için değil, 8192 baytlık tamponu boşaldıkça read yapmaktadır.
                // Bir read bazen bir satır, bazen birkaç satır, bazen de yarım satır getirmektedir.
                // İstemci bağlantıyı kapatınca read 0 döndürmekte ve readLine null vermektedir; bu döngü
                // bunu denetlemediği için null satırları sonsuza kadar basmaktadır.
                String request = in.readLine();
                // → write(1, "Server, received this message from Socket: ...\n", 2066)
                System.out.println("Server, received this message from Socket: " + request);
                lock.unlock();
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}
