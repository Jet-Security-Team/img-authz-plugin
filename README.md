
![scheme](https://github.com/Jet-Security-Team/img-authz-plugin/blob/main/img/scheme.png)


# Docker Image Authorization Plugin

Данный Authz-плагин используется для проверки цифровых подписей образов контейнеров и является форком и адаптацией исходных проектов crosslibs/img-authz-plugin и SixSq/img-authz-plugin.

Для обеспечения и проверки рабочих подписей образов контейнеров используется внешний сервис Notary.

Для дополнительной информации обратитесь к документации Docker по плагинам или оригинальным репозиториям плагина:
https://github.com/crosslibs/img-authz-plugin и https://github.com/SixSq/img-authz-plugin.

Плагин был протестирован и работает в том числе на Alt Linux.

## Сборка и установка плагина

 1. Git clone https://github.com/Jet-Security-Team/img-authz-plugin/ && cd img-authz-plugin

 2. Сборка плагина:

    ```bash
    make build
    ```

    На данном шаге происходит сборка базового образа Docker, содержащего скомпилированный код плагина и необходимые зависимости. Создается промежуточный контейнер (staging container), из которого извлекаются скомпилированные файлы в директорию сборки плагина. Более подробно - [Docker Plugin documentation](https://docs.docker.com/engine/extend/#developing-a-plugin)

 2. Установка плагина:

    ```bash
    make create
    ```
    
    Данная команда удалит предыдущую установку плагина, если их названия совпадают. Перед выполнением данной команды необходимо убедиться, что, демон Docker не запущен с активированным плагином (то есть параметр --authorization-plugin не установлен).
    
    После выполнения данного шага у вас появится новый плагин Docker, однако он будет отключён.

 3. Публикация плагина в регистри (опциально):

    ```bash
    make push
    ```
    
    Данная команда запушит плагин в указанный Docker registry. **Необходимо убедиться**, что были указаны учетные данные для доступа к регистри перед выполнением данного шага (`docker login ...`), а также внесены изменения в параметры файла Makefile - PLUGIN_NAME

---

 Вместо последовательного выполнения всех команд, указанных в данном разделе, можно выполнить единую команду `make all`


## Настройка плагина

Для настройки плагина требуются следующие параметры: `REGISTRY`, `NOTARY` и `NOTARY_ROOT_CA`. Если параметры не указываны, плагин использует следующие настройки по умолчанию - `docker.io` and `notary.docker.io`.

 - `REGISTRY` - регистри, где хранятся образы, используемые Docker в формате _host:port_. Например, REGISTRY="reg.local"
 - `NOTARY` - адрес Notary-сервера, который хранит данные о подписях образов в формате _https://fqdn:port_. Например, NOTARY="https://reg.local:4443"
 - `NOTARY_ROOT_CA` - сертификат, используемый Notary-сервисом. Например, NOTARY_ROOT_CA="$(cat /root/.docker/tls/reg.local:4443/root-ca.crt)" 
 
Плагин является специфичным для процессорной архитектуры. Далее <YOUR_CPU_ARCH> обозначает архитектуру процессора целевого устройства. 

### Настройка собранного из исходного кода плагина

 1. `'docker plugin set jet-isc/img-authz-plugin:x86_64 REGISTRY=<registry> NOTARY=<notary-server> NOTARY_ROOT_CA="$(cat .crt path)"` - настройка плагина. До выполнения команды необходимо выполнить шаги - `make build` и `make create`
    Пример всей команды: docker plugin set jet-isc/img-authz-plugin:x86_64 REGISTRY=reg.local NOTARY=https://reg.local:4443 NOTARY_ROOT_CA="$(cat /root/.docker/tls/reg.local:4443/root-ca.crt)"

 3. Добавление опции в файл конфигурации Docker-daemon `/etc/docker/daemon.json`:

```json
"authorization-plugins": ["jet-isc/img-authz-plugin:<YOUR_CPU_ARCH>"]
```

 4. `docker plugin enable jet-isc/img-authz-plugin:<YOUR_CPU_ARCH>` - включение плагина
   
 5. перезапуск docker-daemon `systemctl restart docker`

### Дополнительный вариант настройки с использованием Docker registry (только для архитектур x86_64 и aarch64)

 1. `docker plugin install jet-isc/img-authz-plugin:<YOUR_CPU_ARCH> REGISTRY=<registry> NOTARY=<notary-server> NOTARY_ROOT_CA="$(cat .crt path)"` - установка плагина из Docker registry

ПРИМЕЧАНИЕ: В процессе настройки плагина ему необходимо будет предоставить доступы к различным функциям Docker (например, доступ к сети хоста). В случае, если доступ к необходимому функционалу не был предоставлен, плагин не будет работать корректно. Для автоматического одобрения всех запросоов можно выполнить команду, указанную выше, с параметром --grant-all-permissions.

 2. Добавление опции в файл конфигурации Docker-daemon `/etc/docker/daemon.json`:

```json
"authorization-plugins": ["jet-isc/img-authz-plugin:<YOUR_CPU_ARCH>"]
```

 3. перезапуск docker-daemon `systemctl restart docker`


## Проверка корректности работы плагина

Для тестирования плагина после установки и настройки необходимо выполнить следующую команду:

```bash
docker run --rm -v $(pwd)/test:/tmp -v /var/run/docker.sock:/var/run/docker.sock docker:dind sh -c 'apk update && apk add shunit2 && SHUNIT_COLOR="always" shunit2 /tmp/tests.sh && docker ps' 
```

Вывод команды выглядит следующим образом:

```

██████╗ ██╗   ██╗███╗   ██╗███╗   ██╗██╗███╗   ██╗ ██████╗     ████████╗███████╗███████╗████████╗███████╗
██╔══██╗██║   ██║████╗  ██║████╗  ██║██║████╗  ██║██╔════╝     ╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝██╔════╝
██████╔╝██║   ██║██╔██╗ ██║██╔██╗ ██║██║██╔██╗ ██║██║  ███╗       ██║   █████╗  ███████╗   ██║   ███████╗
██╔══██╗██║   ██║██║╚██╗██║██║╚██╗██║██║██║╚██╗██║██║   ██║       ██║   ██╔══╝  ╚════██║   ██║   ╚════██║
██║  ██║╚██████╔╝██║ ╚████║██║ ╚████║██║██║ ╚████║╚██████╔╝       ██║   ███████╗███████║   ██║   ███████║██╗██╗██╗
╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚═╝╚═╝  ╚═══╝ ╚═════╝        ╚═╝   ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚═╝╚═╝╚═╝


test_docker_pull_is_allowed
test_docker_run_is_allowed
test_pull_is_not_allowed_when_registry_is_not_authorized
test_run_is_not_allowed_when_registry_is_not_authorized
test_pull_is_allowed_when_registry_is_authorized
test_run_is_allowed_when_registry_is_authorized
test_pull_is_allowed_when_tag_not_specified
test_run_is_allowed_when_tag_not_specified

Ran 8 tests.

OK
```


## Обновление настроек плагина

 1. `docker plugin disable jet-isc/img-authz-plugin:<YOUR_CPU_ARCH>` - отключение плагина

 2. `docker plugin set jet-isc/img-authz-plugin:<YOUR_CPU_ARCH> REGISTRY=<registry> NOTARY=<notary-server> NOTARY_ROOT_CA="$(cat .crt path)"` - изменение настроек плагина

 3. `docker plugin enable jet-isc/img-authz-plugin:<YOUR_CPU_ARCH>` - включение плагина

 4. перезапуск docker-daemon `systemctl restart docker`


## Траблшутинг плагина

Логи плагина добавляются к журналам демона Docker, поэтому их можно найти в соответствующем журнале Docker (например, в Ubuntu для этого можно использовать команду journalctl -u docker).


## Удаление Docker Image Authorization Plugin

_(имя плагина в примере - jet-isc/img-authz-plugin:<YOUR_CPU_ARCH>)_

Деактивация плагина
 1. `docker plugin disable jet-isc/img-authz-plugin:<YOUR_CPU_ARCH>` - отключение плагина

 2. Удаление настройки `authorization-plugins`из файла конфигурации /etc/docker/daemon.json
    
 3. перезапуск docker-daemon `systemctl restart docker`
 
 4. `docker plugin rm -f jet-isc/img-authz-plugin:<YOUR_CPU_ARCH>` - удаление плагина
