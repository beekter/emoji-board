# Troubleshooting / Устранение проблем

## "permission denied" при makepkg

Если вы получаете ошибку:
```
go: github.com/beekter/emoji-board/pkg: open .../pkg: permission denied
```

**Решение:**

1. Удалите старую директорию `pkg/` (артефакт от makepkg):
   ```bash
   rm -rf pkg/
   ```

2. Удалите директорию `vendor/` если она есть (артефакт от go mod vendor):
   ```bash
   rm -rf vendor/
   ```

3. Убедитесь что у вас последняя версия:
   ```bash
   git pull origin copilot/rewrite-to-wails
   ```

4. Пересоберите:
   ```bash
   makepkg -si
   ```

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

## Очистка всех артефактов сборки

```bash
make clean
rm -rf pkg/ vendor/
git clean -fdx  # ВНИМАНИЕ: удалит все неотслеживаемые файлы!
```
