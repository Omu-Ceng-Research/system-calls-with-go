# 2023-24 dönemi örneği. Kodun kendisi değiştirilmemiştir; yalnızca yorumlar eklenmiştir.
#
# '→' ile başlayan yorumlar, altındaki satırın yaptığı sistem çağrılarını göstermektedir.
# Bunlar programın strace çıktısından alınmıştır (Python 3, Linux):
#
#     strace -f -e trace=network,read,write python3 -u server.py
#
# TCP sunucusundan (05) farkı: listen ve accept yoktur; tek soket bütün istemcilere hizmet
# vermektedir. Her paketin kimden geldiği recvfrom'un doldurduğu adresten öğrenilmektedir.

from socket import * 
    
serverPort = 12000 
# → socket(AF_INET, SOCK_DGRAM|SOCK_CLOEXEC, IPPROTO_IP) = 3
serverSocket = socket (AF_INET, SOCK_DGRAM)
# → bind(3, {sin_port=htons(12000), sin_addr=inet_addr("0.0.0.0")}, 16) = 0
serverSocket.bind (('', serverPort)) 
print("Sunucu almak icin hazir") 

while True: 
    # → recvfrom(3, "aaaaaaaaaa", 2048, 0, {sin_port=htons(59295), sin_addr=inet_addr("127.0.0.1")}, [16]) = 10
    # Paket gelene kadar uyumaktadır. Gönderenin adresi (ip, port) clientAddress'e yazılmaktadır.
    message, clientAddress = serverSocket.recvfrom(2048) 
    modifiedMessage = message.decode().upper() 
    # → sendto(3, "AAAAAAAAAA", 10, 0, {sin_port=htons(59295), sin_addr=inet_addr("127.0.0.1")}, 16) = 10
    serverSocket.sendto(modifiedMessage.encode(),clientAddress)
    # → write(1, "AAAAAAAAAA", 10) + write(1, "\n", 1)
    print (modifiedMessage)
