# strace ile Sistem Çağrılarını İzlemek

Bu doküman, bir programın çekirdekten ne istediğini gösteren `strace` aracını örneklerle
anlatmaktadır. Buradaki bütün çıktılar, derste kullanılan Docker ortamında (`sysprog-go`
imajı, strace 6.13) komutlar gerçekten çalıştırılarak alınmıştır. Uzun çıktılar `...` ile
kısaltılmıştır; adresler, PID'ler ve süreler her çalıştırmada farklı olacaktır.

## İçindekiler

1. [strace nedir?](#1-strace-nedir)
2. [Nerede çalışır?](#2-nerede-çalışır)
3. [İlk deneme](#3-ilk-deneme)
4. [Bir satırı okumak](#4-bir-satırı-okumak)
5. [Çıktıyı dosyaya yazmak: -o](#5-çıktıyı-dosyaya-yazmak--o)
6. [Süzmek: -e trace, -P, -Z](#6-süzmek--e-trace--p--z)
7. [Saymak: -c](#7-saymak--c)
8. [Zaman: -t, -tt, -r, -T](#8-zaman--t--tt--r--t)
9. [Veriyi görmek: -s, -x, -y, -e write](#9-veriyi-görmek--s--x--y--e-write)
10. [Çocuk process'ler ve thread'ler: -f, -ff](#10-çocuk-processler-ve-threadler--f--ff)
11. [Çalışan bir programa bağlanmak: -p](#11-çalışan-bir-programa-bağlanmak--p)
12. [Sinyaller](#12-sinyaller)
13. [Çağrıyı kim yaptı? -k](#13-çağrıyı-kim-yaptı--k)
14. [Go programlarını izlemek](#14-go-programlarını-izlemek)
15. [Python ve Java programlarını izlemek](#15-python-ve-java-programlarını-izlemek)
16. [Tarifler: soru → komut](#16-tarifler-soru--komut)
17. [Sık yapılan hatalar](#17-sık-yapılan-hatalar)
18. [Hızlı başvuru](#18-hızlı-başvuru)

---

## 1. strace nedir?

Bir program dosya açmak, ağa veri göndermek, bellek almak, başka bir program başlatmak
gibi işleri kendisi yapamamaktadır; bunları çekirdekten **sistem çağrısı** (system call)
ile istemektedir. `strace`, bir programın yaptığı her sistem çağrısını; argümanlarıyla,
dönüş değeriyle ve hata koduyla birlikte ekrana yazmaktadır.

```
strace ./program        →   openat("veri.txt", O_RDONLY) = 3
                            read(3, "merhaba\n", 4096) = 8
                            write(1, "merhaba\n", 8) = 8
                            close(3) = 0
```

strace bunu programı değiştirmeden, çekirdeğin `ptrace(2)` mekanizmasıyla yapmaktadır:
çekirdek, izlenen program her sistem çağrısına girdiğinde ve çıktığında onu durdurmakta,
strace'e haber vermektedir. strace argümanları okuyup yazdıktan sonra programın devam
etmesine izin vermektedir.

Bunun iki sonucu vardır:

- Programın kaynak koduna ihtiyaç yoktur. C, Go, Python, Java ya da `ls` gibi hazır bir
  program; hepsi aynı biçimde izlenebilmektedir.
- strace programı **yavaşlatmaktadır**. Her çağrıda iki kez durup devam etmek gerekir.
  Bu yüzden strace altında yapılan süre ölçümleri gerçeği yansıtmamaktadır (bkz. bölüm 8).

## 2. Nerede çalışır?

strace **yalnızca Linux'ta** çalışmaktadır.

| Sistem | Durum |
|---|---|
| Ders ortamı (Docker) | Kurulu gelmektedir. `./baslat.sh` ya da `baslat.cmd` ile açınız. |
| Ubuntu / Debian | `sudo apt install strace` |
| Windows | Doğrudan yoktur. Docker ortamını ya da WSL2 içindeki Ubuntu'yu kullanınız. |
| macOS | Doğrudan yoktur. Docker ortamını kullanınız (macOS'ta benzeri `dtruss`'tır). |

Belgesi: `man strace`. Bir sistem çağrısının belgesi: `man 2 openat`, `man 2 read`.

## 3. İlk deneme

```bash
strace echo merhaba
```

```
execve("/usr/bin/echo", ["echo", "merhaba"], 0x7ffedee95d68 /* 12 vars */) = 0
brk(NULL)                               = 0x6112bcf6a000
mmap(NULL, 8192, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0) = 0x711d67fb1000
access("/etc/ld.so.preload", R_OK)      = -1 ENOENT (No such file or directory)
openat(AT_FDCWD, "/etc/ld.so.cache", O_RDONLY|O_CLOEXEC) = 3
fstat(3, {st_mode=S_IFREG|0644, st_size=10611, ...}) = 0
mmap(NULL, 10611, PROT_READ, MAP_PRIVATE, 3, 0) = 0x711d67fae000
close(3)                                = 0
openat(AT_FDCWD, "/lib/x86_64-linux-gnu/libc.so.6", O_RDONLY|O_CLOEXEC) = 3
read(3, "\177ELF\2\1\1\3\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0p\236\2\0\0\0\0\0"..., 832) = 832
...
mprotect(0x711d67f9b000, 16384, PROT_READ) = 0
...
getrandom("\xe0\x64\xe8\x20\x67\x77\xbd\x1e", 8, GRND_NONBLOCK) = 8
brk(NULL)                               = 0x6112bcf6a000
brk(0x6112bcf8b000)                     = 0x6112bcf8b000
fstat(1, {st_mode=S_IFREG|0644, st_size=2253, ...}) = 0
write(1, "merhaba\n", 8merhaba
)                = 8
close(1)                                = 0
close(2)                                = 0
exit_group(0)                           = ?
+++ exited with 0 +++
```

Basit bir `echo` için ~40 satır görülmektedir. Bunlar üç gruba ayrılmaktadır:

| Satırlar | Ne oluyor? |
|---|---|
| `execve` | Kabuk, `echo` programını çalıştırmaktadır. İlk satır her zaman budur. |
| `openat(... libc.so.6)`, `mmap`, `mprotect` | Dinamik yükleyici C kütüphanesini (libc) belleğe eşlemektedir. Programın kendi işi değildir. |
| `write(1, "merhaba\n", 8)` | **Programın asıl işi.** 1 numaralı dosya tanımlayıcısına (stdout) 8 bayt yazılmaktadır. |
| `exit_group(0)` | Program 0 koduyla çıkmaktadır. |

`write` satırının ortasında görülen `merhaba` yazısı programın kendi çıktısıdır: strace
kendi çıktısını stderr'e, program ise stdout'a yazdığı için ikisi ekranda karışmaktadır.
Bu yüzden çoğu zaman `-o` kullanılmaktadır (bölüm 5).

## 4. Bir satırı okumak

```
openat(AT_FDCWD, "/tmp/not.txt", O_RDONLY) = 3
└──┬─┘ └─────────────┬──────────────────┘   └┬┘
 çağrı           argümanlar               dönüş değeri
```

| Görünüm | Anlamı | Örnek |
|---|---|---|
| `= 3` | Başarılı. Değerin anlamı çağrıya göre değişir: `openat` için yeni dosya tanımlayıcısı, `read`/`write` için bayt sayısı. | `read(3, "merhaba\n", 262144) = 8` |
| `= -1 ENOENT (No such file or directory)` | Başarısız. `errno` adı ve açıklaması yazılır. | `openat(AT_FDCWD, "/yok/dosya.txt", O_RDONLY) = -1 ENOENT (...)` |
| `= ?` | Çağrı geri dönmemektedir. | `exit_group(0) = ?` |
| `"merhaba\n"` | Bellekteki veri, C dizgisi biçiminde. Varsayılan olarak 32 karakterden sonrası `...` ile kesilir. | `read(3, "\177ELF\2\1\1..."..., 832)` |
| `\304\261` | Yazdırılamayan baytlar sekizlik (octal) yazılır. Bu ikisi UTF-8'de "ı" harfidir. | `"k\304\261rp\304\261"` |
| `O_RDONLY\|O_CLOEXEC` | Bayraklar (flag) adlarıyla, `\|` ile birleştirilmiş yazılır. | `SOCK_STREAM\|SOCK_NONBLOCK` |
| `{st_mode=..., st_size=8, ...}` | Bir yapı (struct). | `fstat(3, {st_mode=S_IFREG\|0644, ...})` |
| `0x7ffd...` | Bir bellek adresi (işaretçi). | `brk(NULL) = 0x6112bcf6a000` |
| `/* 12 vars */` | strace'in kısalttığı uzun bir dizi. | `execve(..., 0x7ffe... /* 12 vars */)` |
| `<unfinished ...>` ve `<... read resumed>` | Çağrı başka bir thread'in çağrısıyla bölünerek yazılmıştır (bölüm 10). | |
| `[pid  1162]` | Çağrıyı yapan process/thread (yalnızca `-f` ile). | |
| `--- SIGCHLD {...} ---` | Programa bir sinyal gelmiştir. Sistem çağrısı değildir. | |
| `+++ exited with 0 +++` | Program çıkmıştır (ya da `+++ killed by SIGTERM +++`). | |

Dosya tanımlayıcıları (file descriptor, fd) programın açtığı dosya ve soketlerin
numaralarıdır. 0, 1 ve 2 her programda hazır gelmektedir:

| fd | Adı | Ne |
|---|---|---|
| 0 | stdin | klavye ya da `<` ile verilen dosya |
| 1 | stdout | ekran ya da `>` ile verilen dosya |
| 2 | stderr | hata çıktısı; strace de kendi çıktısını buraya yazar |
| 3, 4, ... | | programın açtığı dosyalar ve soketler, sırayla |

Bir hata satırı örneği:

```bash
strace -e trace=openat cat /yok/dosya.txt
```

```
openat(AT_FDCWD, "/etc/ld.so.cache", O_RDONLY|O_CLOEXEC) = 3
openat(AT_FDCWD, "/lib/x86_64-linux-gnu/libc.so.6", O_RDONLY|O_CLOEXEC) = 3
openat(AT_FDCWD, "/yok/dosya.txt", O_RDONLY) = -1 ENOENT (No such file or directory)
cat: /yok/dosya.txt: No such file or directory
+++ exited with 1 +++
```

`cat`'in ekrana yazdığı hata mesajının kaynağı, bir satır yukarıdaki `ENOENT`'tir.

## 5. Çıktıyı dosyaya yazmak: -o

```bash
strace -o iz.txt cat /etc/hostname
cat iz.txt
grep openat iz.txt
```

Programın çıktısı ekranda, strace'in çıktısı `iz.txt` dosyasında kalmaktadır; ikisi
karışmamaktadır. Uzun çıktılar dosyada `grep`, `less` ve `wc -l` ile rahatça
incelenebilmektedir.

strace çıktısını dosya yerine doğrudan `grep`'e vermek için stderr'in yönlendirilmesi
gerekmektedir, çünkü strace stderr'e yazmaktadır:

```bash
strace cat /etc/hostname 2>&1 | grep openat      # doğru
strace cat /etc/hostname | grep openat           # yanlış: grep yalnızca cat'in çıktısını görür
```

## 6. Süzmek: -e trace, -P, -Z

### Belirli çağrılar

```bash
strace -e trace=openat,read,write,close cat /tmp/not.txt
```

```
openat(AT_FDCWD, "/etc/ld.so.cache", O_RDONLY|O_CLOEXEC) = 3
close(3)                                = 0
openat(AT_FDCWD, "/lib/x86_64-linux-gnu/libc.so.6", O_RDONLY|O_CLOEXEC) = 3
read(3, "\177ELF\2\1\1\3\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0p\236\2\0\0\0\0\0"..., 832) = 832
close(3)                                = 0
openat(AT_FDCWD, "/tmp/not.txt", O_RDONLY) = 3
read(3, "merhaba\n", 262144)            = 8
write(1, "merhaba\n", 8)                = 8
read(3, "", 262144)                     = 0
close(3)                                = 0
close(1)                                = 0
close(2)                                = 0
+++ exited with 0 +++
```

`cat`'in işi açıkça görülmektedir: dosyayı açmak (fd 3), 256 KB'lık parçalarla okumak,
okuduğunu stdout'a yazmak, `read` 0 döndürünce (dosya sonu) kapatmak.

### Çağrı sınıfları

Tek tek ad yazmak yerine çağrı sınıfları kullanılabilmektedir:

| Sınıf | İçerdiği çağrılar (örnekler) |
|---|---|
| `%file` | Dosya adı alan çağrılar: `openat`, `stat`, `statx`, `access`, `unlink`, `rename` |
| `%desc` | Dosya tanımlayıcısı alan çağrılar: `read`, `write`, `close`, `fstat`, `lseek`, `fsync` |
| `%network` (ya da `network`) | Soket çağrıları: `socket`, `bind`, `listen`, `accept4`, `connect`, `sendto`, `recvfrom`, `setsockopt` |
| `%process` | Process yaşamı: `execve`, `clone`, `clone3`, `fork`, `wait4`, `exit_group`, `kill` |
| `%signal` | Sinyal çağrıları: `rt_sigaction`, `rt_sigprocmask`, `kill` |
| `%memory` | Bellek: `mmap`, `munmap`, `mprotect`, `brk`, `madvise` |
| `%ipc` | System V IPC: `shmget`, `msgsnd`, `semop` |

```bash
strace -e trace=%file ls /tmp/not.txt
```

```
execve("/usr/bin/ls", ["ls", "/tmp/not.txt"], 0x7ffdaa276288 /* 12 vars */) = 0
access("/etc/ld.so.preload", R_OK)      = -1 ENOENT (No such file or directory)
openat(AT_FDCWD, "/etc/ld.so.cache", O_RDONLY|O_CLOEXEC) = 3
openat(AT_FDCWD, "/lib/x86_64-linux-gnu/libselinux.so.1", O_RDONLY|O_CLOEXEC) = 3
...
statx(AT_FDCWD, "/tmp/not.txt", AT_STATX_SYNC_AS_STAT|AT_NO_AUTOMOUNT, STATX_MODE, {...}) = 0
/tmp/not.txt
+++ exited with 0 +++
```

Bir ağ programının yalnızca ağ ve okuma/yazma çağrıları:

```bash
strace -f -e trace=network,read,write ./d05 istemci
```

### Bir çağrıyı dışarıda bırakmak: `!`

Ünlem işareti, listedeki çağrıların **dışındakileri** göstermektedir. Kabuğun `!`
işaretini yorumlamaması için tek tırnak kullanınız:

```bash
strace -f -e trace='!futex,nanosleep,mmap,rt_sigaction,rt_sigprocmask' ./d05 istemci
```

### Yalnızca başarısız çağrılar: -Z (ya da -e status=failed)

"Program neden çalışmıyor?" sorusunun ilk adımı genellikle budur:

```bash
strace -Z python3 -c 'import json'
```

```
access("/etc/ld.so.preload", R_OK)      = -1 ENOENT (No such file or directory)
openat(AT_FDCWD, "/usr/lib/locale/locale-archive", O_RDONLY|O_CLOEXEC) = -1 ENOENT (...)
newfstatat(AT_FDCWD, "/go/bin/python3", 0x7fffa7dd8de0, 0) = -1 ENOENT (...)
newfstatat(AT_FDCWD, "/usr/local/go/bin/python3", 0x7fffa7dd8de0, 0) = -1 ENOENT (...)
newfstatat(AT_FDCWD, "/usr/local/sbin/python3", 0x7fffa7dd8de0, 0) = -1 ENOENT (...)
...
newfstatat(AT_FDCWD, "/usr/lib/python3.13/encodings/__init__.abi3.so", 0x7fffa7dd8ac0, 0) = -1 ENOENT (...)
...
```

Python'un kendini ve modüllerini birçok yerde aradığı görülmektedir; bunların çoğu
normaldir. Gerçek bir hatayı ararken listenin **sonuna** yakın, programın ekrana yazdığı
hata mesajından hemen önceki başarısız çağrıya bakınız. `-z` ise yalnızca başarılı
çağrıları göstermektedir.

### Bir dosyaya dokunan çağrılar: -P

```bash
strace -P /tmp/not.txt cat /tmp/not.txt /etc/hostname
```

```
openat(AT_FDCWD, "/tmp/not.txt", O_RDONLY) = 3
fstat(3, {st_mode=S_IFREG|0644, st_size=8, ...}) = 0
fadvise64(3, 0, 0, POSIX_FADV_SEQUENTIAL) = 0
copy_file_range(3, NULL, 1, NULL, 9223372035781033984, 0) = -1 EXDEV (Invalid cross-device link)
read(3, "merhaba\n", 262144)            = 8
merhaba
read(3, "", 262144)                     = 0
close(3)                                = 0
064a7c729df5
+++ exited with 0 +++
```

İkinci dosyaya (`/etc/hostname`) ait çağrılar gösterilmemiştir; strace, yalnızca
`/tmp/not.txt` yoluna ve ondan açılan fd 3'e dokunan çağrıları süzmüştür.

## 7. Saymak: -c

`-c`, çağrıları tek tek yazmak yerine program bitince bir özet tablosu basmaktadır:

```bash
strace -c ls /
```

```
% time     seconds  usecs/call     calls    errors syscall
------ ----------- ----------- --------- --------- ----------------
 19.44    0.000168          84         2           getdents64
 17.59    0.000152           6        22           mmap
 11.57    0.000100          14         7           openat
...
```

| Sütun | Anlamı |
|---|---|
| `% time` | Toplam çekirdek süresinin yüzde kaçı bu çağrıda geçti |
| `seconds` | Bu çağrıda geçen toplam süre |
| `usecs/call` | Çağrı başına ortalama süre (mikrosaniye) |
| `calls` | Kaç kez çağrıldı |
| `errors` | Kaçı hata döndürdü |

Sıralamayı değiştirmek için `-S`: `-S calls` (çağrı sayısı), `-S time`, `-S name`.

```bash
strace -c -S calls ls -R /usr/share/doc > /dev/null
```

```
% time     seconds  usecs/call     calls    errors syscall
------ ----------- ----------- --------- --------- ----------------
 35.61    0.009459          17       544           getdents64
  7.16    0.001901           6       280           close
  6.58    0.001748           6       279           fstat
 40.43    0.010739          38       278           openat
  8.79    0.002335           8       273           statx
...
------ ----------- ----------- --------- --------- ----------------
100.00    0.026562          15      1726        10 total
```

`-c` ile süzme birlikte kullanılabilmektedir: `-c -Z` yalnızca hataları sayar.

### Örnek: tamponlama bir programın çağrı sayısını nasıl değiştirir?

Aynı Python programı 1000 satır yazmaktadır; tek fark `flush=True`:

```bash
python3 -c 'for i in range(1000): print(i)' > /dev/null              # tamponlu
python3 -c 'for i in range(1000): print(i, flush=True)' > /dev/null  # her satırda boşalt
```

`strace -c -e trace=write` ile:

```
tamponlu:        1 write
flush=True:   1000 write
```

Tamponlu hâlde Python satırları bellekte biriktirmekte ve tek bir `write` ile
yazmaktadır. Her `write` bir sistem çağrısı, yani çekirdeğe bir geçiştir. Ödevdeki
M2 ölçümü (fsync'li ve fsync'siz yazma) bu tür bir karşılaştırmadır.

## 8. Zaman: -t, -tt, -r, -T

| Seçenek | Satırın başına / sonuna eklenen |
|---|---|
| `-t` | saat (saniye hassasiyetinde) |
| `-tt` | saat, mikrosaniye hassasiyetinde |
| `-ttt` | Unix zamanı (1970'ten beri geçen saniye) |
| `-r` | bir önceki satırdan bu yana geçen süre |
| `-T` | satırın sonunda, **çağrının kendisinin** süresi: `<0.000030>` |

```bash
strace -tt -T -o iz.txt sh -c 'sleep 0.2; echo bitti'
```

```
10:16:49.431424 execve("/usr/bin/sh", ["sh", "-c", "sleep 0.2; echo bitti"], ...) = 0 <0.009685>
10:16:49.441452 brk(NULL)               = 0x5d7ecc2f8000 <0.000022>
10:16:49.441920 openat(AT_FDCWD, "/etc/ld.so.cache", O_RDONLY|O_CLOEXEC) = 3 <0.000056>
...
```

```bash
strace -r -e trace=clock_nanosleep,write,exit_group sh -c 'sleep 0.2; echo bitti'
```

```
     0.000000 --- SIGCHLD {si_signo=SIGCHLD, si_code=CLD_EXITED, si_pid=77, ...} ---
     0.000759 write(1, "bitti\n", 6)    = 6
     0.000140 exit_group(0)             = ?
     0.000465 +++ exited with 0 +++
```

### Örnek: fsync ne kadar sürüyor?

```bash
strace -T -e trace=openat,write,fsync,close python3 -c "
import os
f = open('/tmp/f.txt', 'w')
for i in range(3): f.write('satır %d\n' % i)
f.flush(); os.fsync(f.fileno()); f.close()"
```

```
openat(AT_FDCWD, "/tmp/f.txt", O_WRONLY|O_CREAT|O_TRUNC|O_CLOEXEC, 0666) = 3 <0.000057>
write(3, "sat\304\261r 0\nsat\304\261r 1\nsat\304\261r 2\n", 27) = 27 <0.000024>
fsync(3)                                = 0 <0.004615>
close(3)                                = 0 <0.000020>
```

`write` 24 mikrosaniye, `fsync` 4,6 milisaniye, yani ~190 kat daha uzun sürmüştür.
`write` veriyi yalnızca çekirdeğin belleğine (sayfa önbelleği) bırakmakta, `fsync`
ise diske ulaşmasını beklemektedir. Ödevde "OK bir sözdür" kuralının bedeli budur.

> **Dikkat:** strace programı yavaşlattığı için `-T` ile görülen süreler gerçek
> sürelerden büyüktür. Kısa çağrıların (`write`, `read`) mutlak değerine değil, uzun
> çağrılarla aralarındaki farka bakınız. Asıl performans ölçümü strace olmadan
> yapılmalıdır.

## 9. Veriyi görmek: -s, -x, -y, -e write

### -s: dizgileri kesmeden göstermek

Varsayılan olarak dizgilerin ilk 32 karakteri gösterilmektedir:

```bash
strace -e trace=write echo "Bu uzun bir cümledir ve kırpılacaktır, çünkü varsayılan uzunluk 32 karakterdir."
```

```
write(1, "Bu uzun bir c\303\274mledir ve k\304\261rp\304\261"..., 88) = 88
```

```bash
strace -s 200 -e trace=write echo "Bu uzun bir cümledir ..."
```

```
write(1, "Bu uzun bir c\303\274mledir ve k\304\261rp\304\261lacakt\304\261r, \303\247\303\274nk\303\274 varsay\304\261lan uzunluk 32 karakterdir.\n", 88) = 88
```

### -x ve -xx: onaltılık gösterim

```bash
strace -xx -e trace=write echo "ığ"
```

```
write(1, "\xc4\xb1\xc4\x9f\x0a", 5) = 5
```

"ı" iki bayttır (`c4 b1`), "ğ" iki bayttır (`c4 9f`), sondaki `0a` satır sonudur.
`-x` yalnızca yazdırılamayan baytları, `-xx` bütün baytları onaltılık yazmaktadır.

### -e read=FD ve -e write=FD: tam veri dökümü

```bash
strace -e trace=write -e write=1 echo merhaba
```

```
write(1, "merhaba\n", 8) = 8
 | 00000  6d 65 72 68 61 62 61 0a                           merhaba.         |
```

İkili (binary) protokolleri incelerken, ör. ödevdeki 1536 baytlık vektörler,
kullanışlıdır.

### -y ve -yy: fd numarası yerine ne olduğu

```bash
strace -y -e trace=openat,read,write,close cat /tmp/not.txt > /dev/null
```

```
openat(AT_FDCWD</tmp>, "/tmp/not.txt", O_RDONLY) = 3</tmp/not.txt>
read(3</tmp/not.txt>, "merhaba\n", 262144) = 8
write(1</dev/null>, "merhaba\n", 8)     = 8
read(3</tmp/not.txt>, "", 262144)       = 0
close(3</tmp/not.txt>)                  = 0
```

`-yy`, soketler için protokolü ve bağlantı numarasını da göstermektedir:

```bash
strace -yy -e trace=network ./d05 istemci
```

```
socket(AF_INET, SOCK_STREAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_IP) = 3<TCP:[21750150]>
connect(3<TCP:[21750150]>, {sa_family=AF_INET, sin_port=htons(12001), sin_addr=inet_addr("127.0.0.1")}, 16) = -1 EINPROGRESS (...)
```

## 10. Çocuk process'ler ve thread'ler: -f, -ff

Varsayılan olarak strace yalnızca başlattığı process'i izlemektedir. Bir programın
başlattığı çocuk process'leri ve açtığı thread'leri de izlemek için `-f` gerekmektedir:

```bash
strace -f -e trace=process sh -c 'ls / | wc -l'
```

```
execve("/usr/bin/sh", ["sh", "-c", "ls / | wc -l"], 0x7ffe0c851cf8 /* 12 vars */) = 0
clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|CLONE_CHILD_SETTID|SIGCHLD, ...) = 1162
strace: Process 1162 attached
[pid  1162] execve("/usr/bin/ls", ["ls", "/"], 0x600fecdd6958 /* 12 vars */ <unfinished ...>
[pid  1161] clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|CLONE_CHILD_SETTID|SIGCHLD, ...) = 1163
strace: Process 1163 attached
[pid  1161] wait4(-1,  <unfinished ...>
[pid  1162] <... execve resumed>)       = 0
[pid  1163] execve("/usr/bin/wc", ["wc", "-l"], 0x600fecdd6958 /* 12 vars */) = 0
[pid  1162] exit_group(0)               = ?
[pid  1163] exit_group(0)               = ?
[pid  1162] +++ exited with 0 +++
[pid  1161] <... wait4 resumed>[{WIFEXITED(s) && WEXITSTATUS(s) == 0}], 0, NULL) = 1162
[pid  1163] +++ exited with 0 +++
--- SIGCHLD {si_signo=SIGCHLD, si_code=CLD_EXITED, si_pid=1162, si_uid=1000, ...} ---
wait4(-1, [{WIFEXITED(s) && WEXITSTATUS(s) == 0}], 0, NULL) = 1163
exit_group(0)                           = ?
+++ exited with 0 +++
```

Bir kabuk boru hattının (pipe) çekirdek düzeyindeki hikâyesi budur:

1. Kabuk (`sh`, pid 1161) iki kez `clone` yapmaktadır: pid 1162 ve 1163.
2. 1162 kendini `ls`'e, 1163 kendini `wc`'ye dönüştürmektedir (`execve`).
3. Kabuk `wait4` ile ikisinin bitmesini beklemektedir. Biten çocuk için çekirdek kabuğa
   `SIGCHLD` göndermektedir.

`<unfinished ...>` ve `<... resumed>` satırları, iki process'in çağrılarının aynı anda
sürdüğünü göstermektedir. strace bir çağrının girişini yazdıktan sonra araya başka bir
process'in çağrısı girerse, ilk çağrının sonucunu daha sonra "resumed" satırında
yazmaktadır.

### -ff: her process ayrı dosyaya

```bash
strace -ff -o iz sh -c 'ls / | wc -l'
ls iz.*
```

```
iz.1243  iz.1244  iz.1245
```

Her dosya bir process'in çağrılarını içermektedir; çok thread'li programlarda çıktıyı
okumak böylece kolaylaşmaktadır.

## 11. Çalışan bir programa bağlanmak: -p

Zaten çalışmakta olan bir program, PID'iyle izlenebilmektedir:

```bash
sleep 30 &          # arka planda başlat; kabuk PID'ini yazar
strace -p <PID>
```

```
strace: Process 35 attached
restart_syscall(<... resuming interrupted clock_nanosleep ...>) = ? ERESTART_RESTARTBLOCK (Interrupted by signal)
--- SIGTERM {si_signo=SIGTERM, si_code=SI_USER, si_pid=36, si_uid=0} ---
+++ killed by SIGTERM +++
```

`sleep` bağlandığımız anda `clock_nanosleep` içinde uyumaktaydı. Bir program "takıldı"
gibi göründüğünde ilk sorulacak soru budur: **hangi çağrıda bekliyor?**

Örnek 05'in Go sunucusu, bir istemci beklerken:

```bash
ders 05 sunucu                        # 1. terminal
pgrep d05                             # 2. terminal: PID'i bulunuz
strace -f -p <PID> -e trace=accept4,read,write,epoll_pwait,close
```

```
strace: Process 1026 attached with 5 threads
[pid  1026] epoll_pwait(5, [{events=EPOLLIN, data=0x3cf023b495000002}], 128, -1, NULL, 0) = 1
[pid  1026] accept4(3, {sa_family=AF_INET6, sin6_port=htons(50590), ... "::ffff:127.0.0.1" ...}, ...) = 4
[pid  1026] accept4(3,  <unfinished ...>
[pid  1030] read(4,  <unfinished ...>
[pid  1026] <... accept4 resumed>0x1547c46ccafc, [112], SOCK_CLOEXEC|SOCK_NONBLOCK) = -1 EAGAIN (...)
[pid  1030] <... read resumed>"istanbul", 1024) = 8
[pid  1030] write(1, "127.0.0.1:50590'den gelen mesaj:"..., 44) = 44
[pid  1030] write(1, "\304\260\305\237lem sonucu mesaj: \"\304\260STANBUL"..., 34) = 34
[pid  1030] write(4, "\304\260STANBUL", 9) = 9
[pid  1030] close(4 <unfinished ...>
...
```

Sunucu `epoll_pwait` içinde beklerken bir istemci bağlanmış; `accept4` yeni soket olarak
fd 4'ü döndürmüş, başka bir thread (1030) bu soketten okuyup cevap yazmıştır.

`Ctrl+C` ile strace'ten ayrılınız; izlenen program ölmemekte, çalışmaya devam
etmektedir.

### İzin hatası

```
strace: attach: ptrace(PTRACE_SEIZE, 1221): Operation not permitted
```

Linux, güvenlik nedeniyle bir programın kendi çocuğu olmayan başka bir programa
bağlanmasını kısıtlamaktadır (`/proc/sys/kernel/yama/ptrace_scope`). Çözümler:

- **Ders ortamında:** Başlatma scriptinin (`baslat.sh` / `baslat.cmd`) açtığı **ikinci
  terminal** root yetkisiyle açılmaktadır; `strace -p`'yi oradan çalıştırınız.
- **Kendi Linux'unuzda:** `sudo strace -p <PID>`.
- Programı doğrudan strace altında başlatmak (`strace ./program`) her zaman izinlidir.

## 12. Sinyaller

Sinyaller strace çıktısında `---` ile görünmektedir:

```bash
sleep 30 &
strace -p <PID>          # başka bir terminalden: kill -TERM <PID>
```

```
--- SIGTERM {si_signo=SIGTERM, si_code=SI_USER, si_pid=42, si_uid=0} ---
+++ killed by SIGTERM +++
```

`si_pid=42`, sinyali gönderen process'tir. `+++ killed by SIGTERM +++`, programın
sinyali yakalamadığını ve öldüğünü göstermektedir. Sinyali yakalayan bir program
(örnek 13) ise ölmeden devam etmektedir.

| Seçenek | Etkisi |
|---|---|
| `-e signal=none` | Sinyal satırlarını gizler |
| `-e signal=SIGTERM` | Yalnızca SIGTERM'i gösterir |
| `-q` / `-qq` | "attached/detached" ve çıkış satırlarını gizler |
| `-e trace=none` | Hiçbir çağrıyı göstermez; yalnızca sinyaller kalır |

```bash
strace -qq -e trace=none -e signal=SIGTERM sh -c 'kill -TERM $$'
```

```
--- SIGTERM {si_signo=SIGTERM, si_code=SI_USER, si_pid=1372, si_uid=1000} ---
+++ killed by SIGTERM +++
```

Go programlarında sık görülen `--- SIGURG ... ---` satırları Go runtime'ının kendi iç
sinyalidir (uzun süren goroutine'leri kesmek için); programın davranışıyla ilgisi
yoktur.

## 13. Çağrıyı kim yaptı? -k

`-k`, her çağrı için programın o anki çağrı yığınını (stack trace) yazmaktadır:

```bash
strace -k -e trace=write echo hi
```

```
write(1, "hi\n", 3) = 3
 > /usr/lib/x86_64-linux-gnu/libc.so.6(__write+0x16) [0x1046b6]
 > /usr/lib/x86_64-linux-gnu/libc.so.6(_IO_file_write+0x25) [0x8b315]
 > /usr/lib/x86_64-linux-gnu/libc.so.6(_IO_do_write+0x19) [0x8a419]
 > /usr/lib/x86_64-linux-gnu/libc.so.6(_IO_file_sync+0xa0) [0x8aaf0]
 > /usr/lib/x86_64-linux-gnu/libc.so.6(fflush+0x3e) [0x7e0ae]
 > /usr/bin/echo() [0x6928]
 ...
 > /usr/lib/x86_64-linux-gnu/libc.so.6(exit+0x1a) [0x422fa]
 ...
```

Aşağıdan yukarıya okunmaktadır: `echo` çıkarken (`exit`) C kütüphanesi tamponu
boşaltmış (`fflush`), bu da `__write` fonksiyonunu ve `write` sistem çağrısını
tetiklemiştir. `echo`'nun yazısı, programın sonunda tampon boşaltılırken yazılmaktadır.

## 14. Go programlarını izlemek

### Önce derleyiniz, sonra izleyiniz

```bash
strace -f -c go run ./05-tcp-buyuk-harf istemci      # yanlış
```

```
% time     seconds  usecs/call     calls    errors syscall
------ ----------- ----------- --------- --------- ------------------
 67.08    2.661571         811      3280       482 futex
  7.70    0.305617         200      1525         3 nanosleep
  4.16    0.165073          46      3520        60 newfstatat
...
  0.36    0.014384         224        64           clone
...
```

`go run` önce Go derleyicisini çalıştırmaktadır; sayılan çağrıların neredeyse hepsi
derleyiciye aittir. Doğru yol:

```bash
go build -o d05 ./05-tcp-buyuk-harf          # Docker'da: ders 05 bunu /calisma/d05 olarak yapar
strace -f -c ./d05 istemci
```

```
% time     seconds  usecs/call     calls    errors syscall
------ ----------- ----------- --------- --------- ------------------
 33.67    0.001707          81        21           nanosleep
 19.92    0.001010          84        12         3 futex
  8.95    0.000454           3       114           rt_sigaction
  6.84    0.000347          14        24           mmap
  6.21    0.000315          78         4           clone
  3.89    0.000197          32         6           read
...
  1.87    0.000095          95         1         1 connect
...
  0.71    0.000036          36         1           socket
  0.65    0.000033           6         5           close
  0.61    0.000031           6         5           setsockopt
  0.47    0.000024           4         5           write
...
------ ----------- ----------- --------- --------- ------------------
100.00    0.005070          19       258         4 total
```

### -f gereklidir

Go runtime birden fazla işletim sistemi thread'i kullanmaktadır. `-f` verilmezse
yalnızca ana thread izlenmekte ve çağrıların bir kısmı görünmemektedir.

### Runtime gürültüsü

258 çağrının çoğu programın kendisine değil, Go runtime'ına aittir:

| Çağrı | Neden? |
|---|---|
| `rt_sigaction` (~114 kez) | Runtime açılışta bütün sinyallerin işleyicisini kurmaktadır. |
| `mmap`, `munmap`, `madvise` | Runtime'ın heap ve goroutine yığınları için bellek alması. |
| `clone` | Runtime'ın yeni thread açması. Goroutine sayısıyla ilgili değildir: 52 goroutine açan örnek 07'de ~6 `clone` görülmektedir. Go `clone`, C kütüphanesi (glibc) ve Java ise `clone3` kullanmaktadır. |
| `futex`, `nanosleep` | Thread'lerin birbirini uyutması, uyandırması ve kısa beklemeleri. |
| `epoll_create1`, `epoll_ctl`, `epoll_pwait`, `eventfd2` | Ağ işlemlerini bekleyen "netpoller". |
| `--- SIGURG ---` | Runtime'ın goroutine'leri kesmek için kendine gönderdiği sinyal. |

Bunlar `!` ile dışarıda bırakılınca programın kendi işi ortaya çıkmaktadır:

```bash
strace -f -c -S calls -e trace='!futex,nanosleep,rt_sigaction,rt_sigprocmask,mmap,munmap,madvise,sigaltstack,gettid,clone,mprotect,set_robust_list,rseq,sched_getaffinity,prlimit64,fcntl' ./d05 istemci
```

```
% time     seconds  usecs/call     calls    errors syscall
------ ----------- ----------- --------- --------- ----------------
 24.24    0.000136          19         7         1 read
 12.83    0.000072          14         5           write
  9.63    0.000054          10         5           close
  4.63    0.000026           5         5           setsockopt
...
  2.85    0.000016          16         1           socket
  4.28    0.000024          24         1         1 connect
...
------ ----------- ----------- --------- --------- ----------------
100.00    0.000561          12        46         2 total
```

### EAGAIN bir hata değildir

```bash
strace -f -e trace=network,read,write ./d05 istemci
```

```
socket(AF_INET, SOCK_STREAM|SOCK_CLOEXEC|SOCK_NONBLOCK, IPPROTO_IP) = 3
connect(3, {sa_family=AF_INET, sin_port=htons(12001), sin_addr=inet_addr("127.0.0.1")}, 16) = -1 EINPROGRESS (Operation now in progress)
getsockopt(3, SOL_SOCKET, SO_ERROR, [0], [4]) = 0
getpeername(3, {sa_family=AF_INET, sin_port=htons(12001), sin_addr=inet_addr("127.0.0.1")}, [112 => 16]) = 0
getsockname(3, {sa_family=AF_INET, sin_port=htons(48662), sin_addr=inet_addr("127.0.0.1")}, [112 => 16]) = 0
setsockopt(3, SOL_TCP, TCP_NODELAY, [1], 4) = 0
setsockopt(3, SOL_SOCKET, SO_KEEPALIVE, [1], 4) = 0
...
write(1, "K\303\274\303\247\303\274k harfli c\303\274mleler girini"..., 35) = 35
read(0, "istanbul\n", 4096)       = 9
...
write(3, "istanbul", 8)           = 8
read(3, 0x3e124acca000, 1024)     = -1 EAGAIN (Resource temporarily unavailable)
read(3, "\304\260STANBUL", 1024)  = 9
```

Go soketleri bloklamayan (`SOCK_NONBLOCK`) modda açmaktadır. `connect`'in `EINPROGRESS`
ve `read`'in `EAGAIN` döndürmesi bu yüzdendir: "şu an hazır değil" demektir. Go
goroutine'i bekletmekte, soket hazır olunca (`epoll_pwait`) çağrıyı tekrarlamaktadır.
Python ve Java'da bu satırlar görülmemektedir (bölüm 15).

### Goroutine'ler farklı thread'lerde görünür

Örnek 06'da beş goroutine aynı anda UDP isteği göndermektedir:

```bash
strace -f -e trace=read,write ./d06 coklu -adet 1
```

```
1088  write(3, "eeeeeeeeee", 10 <unfinished ...>
1088  <... write resumed>)              = 10
1088  read(3, "EEEEEEEEEE", 2048)       = 10
1092  write(6, "cccccccccc", 10 <unfinished ...>
1088  write(1, "[e] EEEEEEEEEE\n", 15)  = 15
1092  <... write resumed>)              = 10
1088  write(7, "\1\0\0\0\0\0\0\0", 8 <unfinished ...>
1090  write(8, "dddddddddd", 10 <unfinished ...>
...
1094  write(9, "aaaaaaaaaa", 10 <unfinished ...>
...
```

Her goroutine kendi soketini (fd 3, 4, 6, 8, 9) kullanmakta; çağrılar farklı thread'lerde
(1088, 1090, 1092, ...) görünmektedir. `write(7, "\1\0\0\0\0\0\0\0", 8)` bir `eventfd`'ye
yazmadır: runtime, epoll'da bekleyen thread'i uyandırmaktadır.

## 15. Python ve Java programlarını izlemek

### Python

```bash
strace -f -e trace=network,read,write python3 tcp-server.py
```

```
socket(AF_INET, SOCK_STREAM|SOCK_CLOEXEC, IPPROTO_IP) = 3
bind(3, {sa_family=AF_INET, sin_port=htons(12001), sin_addr=inet_addr("0.0.0.0")}, 16) = 0
listen(3, 1)                     = 0
accept4(3, {sa_family=AF_INET, sin_port=htons(44028), sin_addr=inet_addr("127.0.0.1")}, [16], SOCK_CLOEXEC) = 4
getsockname(4, {sa_family=AF_INET, sin_port=htons(12001), sin_addr=inet_addr("127.0.0.1")}, [128 => 16]) = 0
recvfrom(4, "istanbul", 1024, 0, NULL, NULL) = 8
sendto(4, "ISTANBUL", 8, 0, NULL, 0) = 8
```

- Python soketleri bloklayan modda kullanmaktadır: `accept4` ve `recvfrom` veri gelene
  kadar çekirdekte beklemektedir. EAGAIN ve epoll yoktur.
- `sock.recv(1024)` → `recvfrom(..., NULL, NULL)`, `sock.send(b)` → `sendto(..., NULL, 0)`.
- Python açılışta modüllerini okurken yüzlerce `openat`, `newfstatat` ve `read` yapmaktadır;
  programın kendi çağrıları çıktının sonundadır.
- Çıktı tamponlandığı için `print` satırları `write(1, ...)` olarak geç görünebilmektedir;
  `python3 -u` ile tamponsuz çalıştırınız.

### Java

JVM açılışta binlerce çağrı yapmaktadır (sınıf dosyalarını okumak, GC ve JIT thread'leri
açmak). Programın kendi çağrılarını bulmak için dosyaya yazıp aramak en kolay yoldur:

```bash
javac Server.java
strace -f -o iz.txt java Server
grep -n 9991 iz.txt            # dinleme soketinin açıldığı satırdan itibaren okuyunuz
```

- Java da bloklayan soketler kullanmaktadır (`accept`, `read`).
- Her `new Thread(...).start()` bir `clone3`'tür; 52 thread açan bir Java programında
  `strace -f -c -e trace=clone3` ~82 sonuç vermektedir (geri kalanı JVM'in kendi thread'leri).
- Tamponsuz akışlar çok sayıda küçük çağrı üretmektedir: örnek 08'deki
  `ObjectInputStream`, 86 baytlık tek bir nesneyi 24 ayrı `read` ile okumaktadır.

## 16. Tarifler: soru → komut

| Soru | Komut |
|---|---|
| Program hangi dosyaları açıyor? | `strace -f -e trace=openat ./prog` |
| Program hangi dosyayı bulamıyor? | `strace -f -Z -e trace=%file ./prog` |
| Hangi yapılandırma dosyasını okuyor? | `strace -f -e trace=openat ./prog 2>&1 \| grep -E '\.conf\|\.json'` |
| Hangi adrese/porta bağlanıyor? | `strace -f -e trace=connect ./prog` |
| Hangi portu dinliyor? | `strace -f -e trace=bind,listen ./prog` |
| Ağa ne gönderiyor? | `strace -f -s 500 -e trace=write,sendto,sendmsg ./prog` |
| Program takıldı, neyi bekliyor? | `strace -f -p <PID>` (2. terminalden) |
| fsync gerçekten çağrılıyor mu? | `strace -f -e trace=fsync,fdatasync ./prog` |
| fsync ne kadar sürüyor? | `strace -f -T -e trace=fsync ./prog` |
| Kaç write yapıyor? Tamponlama var mı? | `strace -f -c -e trace=write ./prog` |
| En çok hangi çağrıda zaman geçiyor? | `strace -f -c -S time ./prog` |
| Hangi çocuk process'leri başlatıyor? | `strace -f -e trace=execve ./prog` |
| Çocuğu bekliyor mu (zombi kalıyor mu)? | `strace -f -e trace=clone,clone3,execve,wait4,exit_group ./prog` |
| Hangi sinyalleri alıyor? | `strace -f -e trace=none ./prog` |
| SIGPIPE / EPIPE nerede oluyor? | `strace -f -e trace=write ./prog 2>&1 \| grep -E 'EPIPE\|SIGPIPE'` |
| mmap ile ne kadar bellek eşliyor? | `strace -f -e trace=mmap,munmap,mremap ./prog` |
| Dosyayı hangi ofsetten okuyor? | `strace -f -e trace=pread64,lseek,read ./prog` |
| Bir dosyaya kim, ne yapıyor? | `strace -f -P /yol/dosya ./prog` |
| Bir fd hangi dosya/soket? | `strace -f -yy ...` |

### Ödevle ilgili

| Ödevde | Komut |
|---|---|
| M3: bir düğüm 1000 Kaydet ve 100 Ara işlerken | `strace -f -c -o m3.txt ./dugum ...` (yükü başka terminalden veriniz) |
| Savunma: "Kaydet sırasında fsync'i gösteriniz" | `strace -f -e trace=write,fsync -p <düğüm PID>` |
| Yazma sırası (metin.dat önce, vektor.dat sonra) | `strace -f -y -e trace=write,fsync -p <PID>` |
| Embedder'ın başlatılması ve öldürülmesi | `strace -f -e trace=clone,execve,kill,wait4 ./koordinator ...` |
| Pipe'tan 1536 baytın kaç read'de geldiği | `strace -f -e trace=read -e read=<fd> ./koordinator ...` |
| mmap ve eşlemenin büyütülmesi | `strace -f -e trace=mmap,munmap,mremap,fstat ./dugum ...` |
| Çökme sonrası kurtarmada ftruncate | `strace -f -e trace=ftruncate,fsync ./dugum ...` |

## 17. Sık yapılan hatalar

| Hata | Sonuç | Doğrusu |
|---|---|---|
| `-f` unutmak | Go, Java ve çok thread'li programlarda çağrıların bir kısmı görünmez. | `strace -f ...` |
| `go run` altında strace | Sayılan çağrıların neredeyse hepsi derleyiciye aittir. | Önce `go build`, sonra `strace ./program` |
| `strace prog \| grep x` | grep, strace'in çıktısını değil programınkini görür. | `strace prog 2>&1 \| grep x` ya da `-o` |
| Dizgilerin `...` ile kesilmesi | Gönderilen verinin tamamı görülmez. | `-s 500` |
| `!` işaretini çift tırnakla yazmak | Kabuk `!`'yi geçmiş komut olarak yorumlar. | Tek tırnak: `-e trace='!futex'` |
| strace altında performans ölçmek | Süreler gerçekte olduğundan büyük çıkar. | Ölçümü strace olmadan yapınız; strace ile yalnızca çağrı sayılarına ve oranlarına bakınız. |
| `-p` ile "Operation not permitted" | Başka bir programa bağlanma izni yoktur. | Ders ortamında 2. terminal (root); kendi Linux'unuzda `sudo` |
| `EAGAIN`'i hata sanmak | Go'da her bekleyen okumada görülür. | Bölüm 14 |
| Çıktıyı ekranda okumaya çalışmak | Yüzlerce satır kayar gider. | `-o iz.txt`, ardından `less iz.txt` ve `grep` |

## 18. Hızlı başvuru

| Seçenek | Anlamı |
|---|---|
| `-o DOSYA` | Çıktıyı dosyaya yaz |
| `-f` | Çocuk process'leri ve thread'leri de izle |
| `-ff -o ÖNEK` | Her process'i ayrı dosyaya yaz (`ÖNEK.PID`) |
| `-p PID` | Çalışan bir programa bağlan |
| `-e trace=a,b,c` | Yalnızca bu çağrılar |
| `-e trace='!a,b'` | Bu çağrılar hariç |
| `-e trace=%file`, `%desc`, `%network`, `%process`, `%signal`, `%memory` | Çağrı sınıfları |
| `-e trace=none` | Hiçbir çağrı (yalnızca sinyaller) |
| `-Z` / `-z` | Yalnızca başarısız / başarılı çağrılar |
| `-P YOL` | Yalnızca bu yola dokunan çağrılar |
| `-c` | Özet tablosu |
| `-S calls` / `-S time` | Özet tablosunun sıralaması |
| `-t`, `-tt`, `-ttt` | Saat damgası |
| `-r` | Bir önceki satırdan bu yana geçen süre |
| `-T` | Her çağrının süresi |
| `-s N` | Dizgilerin ilk N karakterini göster (varsayılan 32) |
| `-x` / `-xx` | Onaltılık gösterim |
| `-e read=FD`, `-e write=FD` | Bu fd'deki verinin tam dökümü |
| `-y` / `-yy` | fd'lerin dosya yolunu / soket bilgisini göster |
| `-k` | Her çağrının yığın izi |
| `-e signal=none` | Sinyal satırlarını gizle |
| `-q`, `-qq` | Bilgi satırlarını gizle |

Bu repodaki her örneğin `README.md` dosyasında, o örneğe ait hazır strace komutları
bulunmaktadır.
