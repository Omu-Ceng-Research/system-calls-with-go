# 2023-24 dönemi örneği. Kodun kendisi değiştirilmemiştir; yalnızca yorumlar eklenmiştir.
#
# '→' ile başlayan yorumlar, altındaki satırın yaptığı sistem çağrılarını göstermektedir.
# Bunlar programın strace çıktısından alınmıştır (Python 3, Linux):
#
#     strace -f -e trace=network,read,write python3 tcp-server.py
#
# Go sürümünden (../main.go) farkı: Python soketleri bloklayan (blocking) modda açmaktadır.
# accept ve recv, veri gelene kadar çekirdekte uyumaktadır; Go'daki EAGAIN + epoll_pwait
# döngüsü yoktur. Her Python satırı genellikle tek bir sistem çağrısına karşılık gelmektedir.

from socket import * 

serverPort = 12001 
# → socket(AF_INET, SOCK_STREAM|SOCK_CLOEXEC, IPPROTO_IP) = 3
serverSocket = socket(AF_INET, SOCK_STREAM) 
# → bind(3, {sin_port=htons(12001), sin_addr=inet_addr("0.0.0.0")})
# '' adresi 0.0.0.0'a, yani bütün ağ arayüzlerine karşılık gelmektedir.
# Python, Go'dan farklı olarak SO_REUSEADDR seçeneğini açmamaktadır. Sunucu kapandıktan
# sonra ~60 saniye (TIME_WAIT) bu satır "OSError: [Errno 98] Address already in use"
# hatası verebilmektedir (bkz. ../../12-so-reuseaddr). Çözüm, bind'dan önce:
#     serverSocket.setsockopt(SOL_SOCKET, SO_REUSEADDR, 1)   → setsockopt(3, SOL_SOCKET, SO_REUSEADDR, [1])
serverSocket.bind(('', serverPort)) 
# → listen(3, 1)   kuyrukta en fazla 1 bekleyen bağlantı (Go'da 4096)
serverSocket.listen(1) 
# → write(1, "Sunucu almak icin hazır\n")   (çıktı bir pipe'a yönlendirilmişse
#   Python bunu tamponda bekletmekte, write daha sonra yapılmaktadır)
print("Sunucu almak icin hazır") 


while True: 
    # → accept4(3, {sin_port=htons(44028), sin_addr=inet_addr("127.0.0.1")}, ..., SOCK_CLOEXEC) = 4
    # → getsockname(4)
    # İstemci bağlanana kadar process burada, çekirdekte uyumaktadır. Yeni soket fd 4'tür.
    connectionSocket, addr = serverSocket.accept()
    # → recvfrom(4, "istanbul", 1024, 0, NULL, NULL) = 8
    # Python'un recv'i, read yerine adres alanları boş (NULL) bir recvfrom yapmaktadır.
    # Gelen kadar bayt dönmektedir (8), 1024 değil. decode() bellekte çalışmaktadır.
    sentence = connectionSocket.recv(1024).decode()
    print('Sunucuya gelen mesaj: ', sentence )
    # Sistem çağrısı yoktur. Dikkat: Python'un upper()'ı Türkçe kurallarını bilmemektedir;
    # "istanbul".upper() == "ISTANBUL" olmaktadır (Go sürümünde "İSTANBUL").
    capitalizedSentence = sentence.upper()
    print('İşlem sonucu mesaj: ', capitalizedSentence )
    # → sendto(4, "ISTANBUL", 8, 0, NULL, 0) = 8
    connectionSocket.send(capitalizedSentence.encode()) 
# Bu satır while True döngüsünün DIŞINDADIR ve hiç çalışmamaktadır. Bu yüzden her
# istemcinin soketi (fd 4, 5, 6 ...) açık kalmaktadır. Doğrusu, close'un döngünün içinde
# olmasıdır: → close(4)
connectionSocket.close()
