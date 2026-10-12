# мне лень делать README
## сорян
# build
```bash
go get
go build
```

# clients
Клиенты лежат в `app/` — это Kotlin Multiplatform проект:
- `shared` — общий UI на Compose Multiplatform, `VpnManager`, настройки
- `androidApp` — Android-клиент, Go-часть подключена через `libs/vpn.aar`
- `desktopApp` — GUI-клиент для Windows и Linux. Туннель поднимает тот же Go-бинарь
  в режиме `gui-backend`: GUI запускает его с правами админа (UAC на Windows,
  `pkexec` на Linux), и он подключается обратно к GUI по localhost.

Для desktop нужны JDK 17 и Go (бинарь собирается под текущую ОС автоматически):
```bash
cd app
./gradlew :desktopApp:run            # запустить
./gradlew :desktopApp:packageMsi     # на Windows
./gradlew :desktopApp:packageDeb     # на Linux, или packageRpm
./gradlew :desktopApp:packageArch    # на Arch Linux, нужен makepkg (base-devel)
```
Пакеты появятся в `app/desktopApp/build/compose/binaries/main/`. Собрать их можно
только на той ОС, под которую они делаются, поэтому CI собирает их на своих раннерах.

Android:
```bash
cd app
./gradlew :androidApp:assembleDebug
```
