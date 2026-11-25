# Публикация HoI4 Mod Navigator на GitHub

Готовый репозиторий содержит исходники v2.0, README на двух языках, баннер, GIF-иллюстрацию, шаблоны Issues/PR, автоматическую сборку и черновик релиза. Имя владельца GitHub неизвестно — в командах ниже подставь свой логин. Ничего не пушится автоматически из чата.

## 1. Создай публичный репозиторий

Открой https://github.com/new и выбери:

- Repository name: `HoI4-Mod-Navigator`
- Description: `Context-aware Windows Explorer integration for Hearts of Iron IV modding. Jump between vanilla and mods, copy and compare files.`
- Visibility: Public.
- Не добавляй README, .gitignore или лицензию через GitHub: они уже есть в архиве.

Перед публикацией проверь `LICENSE` (предложена MIT; при согласии укажи предпочитаемое имя правообладателя вместо `HoI4 Mod Navigator contributors`). Не публикуй чужие компоненты без прав. В архиве нет файлов самой игры/модов, настроек FFU2 или персональных путей.

## 2. Первый push

Распакуй исходный архив в удобное место, открой его корневую папку (там, где `go.mod`) и запусти Git Bash. Установи Git for Windows, если его нет. Вводи команды по очереди:

```bash
git init -b main
git add .
git status
git commit -m "Initial public preview of HoI4 Mod Navigator"
git remote add origin https://github.com/YOUR_GITHUB_LOGIN/HoI4-Mod-Navigator.git
git push -u origin main
```

Если Git просит имя/email, настрой их для этого репозитория (`git config user.name "YOUR_NAME"`, `git config user.email "YOUR_EMAIL"`). Для приватности можно взять noreply-адрес в GitHub Settings → Emails. Для входа используй Git Credential Manager/браузер или SSH, а не пароль аккаунта. Не сохраняй токен в репозитории.

После пуша проверь Actions → Build and test: задания `build` и `windows-tests` должны стать зелёными. GitHub Actions использует текущий поддерживаемый Go, собирает DLL и ресурсы из исходников, запускает Go-тесты, C mock-тесты и Windows-тесты на отдельном Windows runner. CI не запускает Explorer с установленным DLL. Если задания падают — не выпускай релиз; приложи лог в Issue. `.github/workflows/build.yml` здесь проверен как конфигурация, но не запущен в твоём GitHub.

## 3. Оформи главную страницу

Справа на главной странице репозитория нажми шестерёнку About и добавь Description из шага 1. Topics (в нижнем регистре, по одному):

`hoi4`, `hearts-of-iron-4`, `modding`, `windows`, `file-explorer`, `shell-extension`, `golang`, `game-modding`, `context-menu`, `utility`

В Settings → General → Social preview загрузи `assets/social-preview.png` (1280×640). Это отдельное изображение для ссылок в Discord, соцсетях и других сервисах. Баннер уже встроен в README. Добавь опциональный статус сборки в README после успешного первого CI: `![Build](https://github.com/YOUR_GITHUB_LOGIN/HoI4-Mod-Navigator/actions/workflows/build.yml/badge.svg)`. После публикации релиза сделай ссылку на `https://github.com/YOUR_GITHUB_LOGIN/HoI4-Mod-Navigator/releases` в блоке Download, заменив логин на свой.

В Settings включи Issues и (если нужен) Discussions, включи Private vulnerability reporting в разделе Security/Advanced Security, если доступен. Включи branch protection/ruleset для main с обязательными проверками CI, если это удобно; не блокируй себе первый push. В Security убедись, что приватный канал отчётов реально доступен. В профиле можно закрепить репозиторий.

## 4. Сними настоящие скриншоты и GIF

В README уже есть аккуратно подписанные иллюстрации и анимированная демонстрация; это не запись твоей Windows. Для настоящей демонстрации:

1. На тестовых папках настрой два проекта (например, `FFU2` и `Example Mod`); проверь все функции из `docs/WINDOWS_TEST_CHECKLIST.md`. Можно создать пустой `hoi4.exe` как маркер в тестовой папке, не запускать его.
2. Установи ScreenToGif с https://www.screentogif.com/ или воспользуйся OBS Studio. Настрой область записи около 1000–1200 px шириной и 10–15 FPS для GIF. Убери личные данные и адреса аккаунтов из кадра. Для скриншотов подойдёт Win+Shift+S.
3. Сними 6–10 секунд: ванилла `gfx/interface` → ПКМ → «Открыть идентичную папку в FFU2» → новое окно мода. Сохрани в `assets/navigation-live.gif`. Вторая запись — ПКМ по одиночному файлу мода → выделение ванильного файла. Сделай `assets/reverse-live.png`, `assets/settings-live.png`.
4. Заменяй ссылки в обоих README на новые файлы, а подписи на «Запись Windows 10/11, версия 2.0». Пока изображения не сняты, не убирай маркировку симуляции. Рекомендуемый GIF до нескольких МБ, без многократных повторов и лишних задержек. Если GIF тяжёлый, оставь PNG и ссылку на видео.
5. `git add assets README.md README_RU.md && git commit -m "Add real Windows demo" && git push`.

## 5. Первый релиз (после проверки)

Сначала пройди чек-лист `docs/WINDOWS_TEST_CHECKLIST.md`. v2.0 работает у тебя по предварительной проверке, но это не полный тест всех сценариев. Поэтому первый публичный релиз лучше обозначить как **2.0.0-beta.1**, а не стабильный 2.0.0. В самом приложении пока отображается v2.0.

```bash
git tag -a v2.0.0-beta.1 -m "HoI4 Mod Navigator 2.0 preview"
git push origin v2.0.0-beta.1
```

В Actions стартует сборка по тегу. Когда она и Windows-тесты пройдут, workflow создаст **черновик prerelease** с EXE, ZIP и SHA256SUMS.txt. В Releases открой черновик, проверь артефакты, вставь описание из `docs/RELEASE_NOTES.md`, оставь галочку pre-release и опубликуй вручную. Если черновик не создан, проверь разрешения GitHub Actions и логи; можно вручную создать release с существующим тегом и скачать файлы из артефакта `windows-x64`. Не ставь отметку stable до проверки реальной Windows.

При скачивании хэши можно сверить в PowerShell:

```powershell
Get-FileHash .\HoI4ModNavigator.exe -Algorithm SHA256
Get-FileHash .\HoI4ModNavigator-Windows-x64.zip -Algorithm SHA256
```

Сравни с файлом SHA256SUMS.txt из релиза. ZIP содержит хэш EXE; внешний SHA256SUMS.txt также содержит хэш ZIP. Хэш проверяет целостность, не заменяет подпись издателя.

## 6. Дальнейшее развитие как продукта

- 2.1: английский интерфейс, прямые переходы между модами, компактное подменю при большом числе модов — это пока идеи, не реализованные функции.
- Добавить результаты реальной проверки Windows 10/11, версию в свойства EXE и при необходимости поддержку DPI.
- Рассмотреть доверенную подпись (например, SignPath Foundation для подходящих открытых проектов) после появления публичного проекта и проверки требований. Подпись не обещает мгновенного исчезновения SmartScreen. Не проси пользователей отключать защиту.
- Не коммить EXE/DLL в Git: выпускай бинарники через Releases, сохраняй исходники и историю. Для следующих релизов не переиспользуй теги; меняй имя DLL при изменении её содержимого, чтобы не перезаписывать загруженное расширение.
- Перед включением обязательной проверки CI убедись, что PR и Actions работают. Секреты подписи и токены — только в GitHub Secrets. Workflow использует major-теги официальных Actions и Dependabot; для более строгой защиты закрепи проверенные SHA коммитов actions.

Справка: [GitHub: создание репозитория](https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-new-repository), [управление релизами](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository), [Go CI](https://docs.github.com/en/actions/tutorials/build-and-test-code/go).
