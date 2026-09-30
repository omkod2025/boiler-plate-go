# Migrations

เพิ่ม Goose SQL migrations ตาม model/policy ด้วย migration role ของ service เท่านั้น
ใช้ `MIGRATION_DB_URL=... MIGRATION_SCHEMA=<owned_schema> service migrate` แยกก่อน deploy
runner ใช้ custom store `<owned_schema>.nxct_schema_migrations` มี prefix ทุก business column และ audit ครบ 6 ตัว
ไม่มี automatic Down และห้าม migration ตอน API boot; expand ก่อน contract หลัง consumers ย้ายครบ
