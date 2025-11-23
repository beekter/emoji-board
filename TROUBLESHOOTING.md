# Troubleshooting / Устранение проблем

## "Wails applications will not build without the correct build tags"

Эта ошибка означает что приложение собрано без необходимых build tags.

**Решение:**

Всегда используйте `-tags desktop,production`:
```bash
go build -tags desktop,production -o emoji-keyboard .
```

Или просто:
```bash
make build
```

## Проблемы с makepkg

Начиная с последней версии, PKGBUILD автоматически очищает все кеши и артефакты перед сборкой.

Если всё равно возникают проблемы:

1. Обновите код:
   ```bash
   git pull origin copilot/rewrite-to-wails
   ```

2. Просто запустите:
   ```bash
   makepkg -si
   ```

PKGBUILD автоматически очистит:
- Старую директорию `pkg/`
- Директорию `vendor/`
- Артефакты сборки `build/`, `emoji-keyboard`, `emoji-board`
- Сгенерированные файлы `frontend/wailsjs/`
- Go кеши сборки

## Ручная очистка (если нужно)

```bash
make clean
rm -rf pkg/ vendor/
```
