# 2023-24 dönemi örneği. Kodun kendisi değiştirilmemiştir; yalnızca yorumlar eklenmiştir.
#
# '→' ile başlayan yorumlar, altındaki satırın yaptığı sistem çağrılarını göstermektedir.
# Bunlar programın strace çıktısından alınmıştır (Python 3, Linux):
#
#     strace -f -e trace=network,read,write,clock_nanosleep python3 -u client2.py
#
# client.py … client5.py yalnızca gönderdikleri harfte ayrılmaktadır (a, b, c, d, e).

import time
from socket import *

serverName = 'localhost'
serverPort = 12000
# → socket(AF_INET, SOCK_DGRAM|SOCK_CLOEXEC, IPPROTO_IP) = 3
# bind çağrılmamaktadır; ilk sendto'da çekirdek boş bir yerel port seçmektedir.
clientSocket = socket (AF_INET, SOCK_DGRAM)

while True:
    message = 'bbbbbbbbb'
    # 'localhost' her turda yeniden çözülmektedir: → connect("/var/run/nscd/socket") = -1 ENOENT ...
    # → sendto(3, "aaaaaaaaaa", 10, 0, {sin_port=htons(12000), sin_addr=inet_addr("127.0.0.1")}, 16) = 10
    # UDP'de connect yoktur; hedef adres her sendto'da ayrıca verilmektedir.
    clientSocket.sendto(message.encode(), (serverName, serverPort))
    # → recvfrom(3, "AAAAAAAAAA", 2048, 0, {sin_port=htons(12000), ...}, [16]) = 10
    # Zaman aşımı yoktur: sunucu kapalıysa ya da paket kaybolduysa istemci burada
    # sonsuza kadar beklemektedir (Go sürümünde SetReadDeadline ile 2 saniye).
    modifiedMessage, serverAddress = clientSocket.recvfrom (2048)
    # → write(1, "AAAAAAAAAA", 10) + write(1, "\n", 1)
    print (modifiedMessage.decode())
    # → clock_nanosleep(CLOCK_MONOTONIC, TIMER_ABSTIME, ...)   1 saniye uyku
    time.sleep(1)
#clientSocket.close()
