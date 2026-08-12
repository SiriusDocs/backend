# Json-схема логов

## Пример успешной операции (`INFO`):
```json
{
  "time": "2026-08-12T10:15:30.123456789Z",
  "level": "INFO",
  "msg": "upload url generated",
  "service": "file_service",
  "env": "prod",
  "source": "file_service.go:42",
  "trace_id": "a1b2c3d4-e5f6-7890-1234-56789abcdef0",
  "user_id": 42,
  "key": "avatars/user_42.png",
  "content_type": "image/png",
  "ttl": "15m0s",
  "result": {
    "status": "ok",
    "duration_ms": 12
  }
}
```

## Пример ошибки (`ERROR`):
```json
{
  "time": "2026-08-12T10:16:05.987654321Z",
  "level": "ERROR",
  "msg": "failed to delete file",
  "service": "file_service",
  "env": "prod",
  "source": "file_service.go:95",
  "trace_id": "c9a2f1b4-8e3d-4c12-9f1a-5b6c7d8e9f0a",
  "user_id": 42,
  "key": "avatars/user_42.png",
  "error": {
    "message": "services.files.DeleteFile: storage.s3.Delete (key=avatars/user_42.png): context deadline exceeded",
    "type": "*fmt.wrapError"
  },
  "result": {
    "status": "error",
    "duration_ms": 3001
  }
}
```