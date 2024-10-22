# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## локальное тестирование

Для запуска локальных тестов с БД необходимо запустить команду 
`docker-compose  -f ./test.docker-compose.yaml up -d && go test ./...`

## запуск линтера

В проекте используется специализированный multichecker. 
Для сборки воспользуйтесь командой `go build ./cmd/staticlint/`
Для запуска - `./staticlint ./...`

## создание сертификатов 

Для создания сертификатов и ключей асимметричного шифрования необходимо запустить команду
go run ./cmd/cert/ -p='server' && go run ./cmd/cert/ -p='client'

## запуск с флагами 
go run ./cmd/server/ -cert './keys/server_cert.pem' -crypto-key './keys/server_privatekey.pem' 
go run ./cmd/agent/ -client-cert './keys/client_cert.pem' -crypto-key './keys/client_privatekey.pem' -server-cert './keys/server_cert.pem'