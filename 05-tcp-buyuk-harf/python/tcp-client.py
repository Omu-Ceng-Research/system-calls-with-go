# 2023-24 dönemi örneği. Kodun kendisi değiştirilmemiştir; yalnızca yorumlar eklenmiştir.
#
# '→' ile başlayan yorumlar, altındaki satırın yaptığı sistem çağrılarını göstermektedir.
# Bunlar programın strace çıktısından alınmıştır (Python 3, Linux):
#
#     strace -f -e trace=network,read,write python3 tcp-client.py

from socket import *

serverName = 'localhost'
serverPort = 12001 

# → socket(AF_INET, SOCK_STREAM|SOCK_CLOEXEC, IPPROTO_IP) = 3
clientSocket = socket (AF_INET, SOCK_STREAM) 
# 'localhost' adı önce IP'ye çevrilmektedir (getaddrinfo):
# → socket(AF_UNIX, ...) + connect("/var/run/nscd/socket") = -1 ENOENT   ad önbelleği yok
# → openat("/etc/hosts") + read(...)                                    localhost = 127.0.0.1
# → connect(3, {sin_port=htons(12001), sin_addr=inet_addr("127.0.0.1")}) = 0
# Bloklayan connect: TCP el sıkışması bitene kadar beklenmektedir (Go'da EINPROGRESS + epoll).
clientSocket.connect((serverName, serverPort)) 

# → write(1, "Kucuk harfli cumleler giriniz:", 30)
# → read(0, "istanbul\n", 8192)    0: stdin (klavye)
sentence = input('Kucuk harfli cumleler giriniz:') 
print('Girilen mesaj: ', sentence) 
print('Encoded mesaj: ', sentence.encode() )

# → sendto(3, "istanbul", 8, 0, NULL, 0) = 8
clientSocket.send(sentence.encode()) 

# → recvfrom(3, "ISTANBUL", 1024, 0, NULL, NULL) = 8   cevap gelene kadar uyumaktadır
modifiedSentence = clientSocket.recv(1024) 
print('Sunucudan gelen mesaj: ', modifiedSentence.decode()) 
# → close(3)
clientSocket.close()
