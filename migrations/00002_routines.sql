-- Routines ที่ runtime เรียกแทนการอ่าน/เขียนตารางตรง:
--   อ่าน (GET/list/lookup) → function  oktf_*  เรียกด้วย SELECT
--   เขียน (INSERT/UPDATE/DELETE) → procedure oktp_* เรียกด้วย CALL, ผลลัพธ์คืนผ่าน OUT parameters
-- procedure ไม่ COMMIT/ROLLBACK เอง: caller คุม transaction (รวมหลายคำสั่งเป็น transaction เดียวได้)
-- ทุก routine ตั้ง search_path และอ้าง object แบบระบุ schema

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION public.oktf_user_get(p_user_profile_id INTEGER)
RETURNS TABLE (user_profile_id INTEGER, email TEXT, full_name TEXT, password_hash TEXT, role TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql STABLE SET search_path = pg_catalog, pg_temp AS $$
  SELECT u.user_profile_id, u.email, u.full_name, u.password_hash, u.role, u.created_at, u.updated_at
  FROM public.okdt_user_profiles u
  WHERE u.user_profile_id = p_user_profile_id
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.oktf_user_credential_get(p_email TEXT)
RETURNS TABLE (user_profile_id INTEGER, password_hash TEXT, role TEXT)
LANGUAGE sql STABLE SET search_path = pg_catalog, pg_temp AS $$
  SELECT u.user_profile_id, u.password_hash, u.role FROM public.okdt_user_profiles u WHERE u.email = p_email
$$;
-- +goose StatementEnd

-- email ซ้ำ → unique_violation (23505) ให้ repository แปลงเป็น domain error
-- +goose StatementBegin
CREATE PROCEDURE public.oktp_user_insert(
  IN p_email TEXT, IN p_full_name TEXT, IN p_password_hash TEXT, IN p_role TEXT,
  OUT o_user_profile_id INTEGER, OUT o_created_at TIMESTAMPTZ, OUT o_updated_at TIMESTAMPTZ)
LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  INSERT INTO public.okdt_user_profiles (email, full_name, password_hash, role)
  VALUES (p_email, p_full_name, p_password_hash, p_role)
  RETURNING user_profile_id, created_at, updated_at INTO o_user_profile_id, o_created_at, o_updated_at;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE PROCEDURE public.oktp_category_insert(
  IN p_category_name TEXT, IN p_user_profile_id INTEGER, IN p_color TEXT, IN p_icon TEXT, IN p_category_type TEXT,
  OUT o_category_id INTEGER, OUT o_created_at TIMESTAMPTZ, OUT o_updated_at TIMESTAMPTZ)
LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  INSERT INTO public.okdt_categories (category_name, user_profile_id, color, icon, category_type)
  VALUES (p_category_name, p_user_profile_id, p_color, p_icon, p_category_type)
  RETURNING category_id, created_at, updated_at INTO o_category_id, o_created_at, o_updated_at;
END $$;
-- +goose StatementEnd

-- แก้เฉพาะ field ที่ส่งมา (NULL = ไม่แก้) ของหมวดหมู่ที่ผู้ใช้เป็นเจ้าของ
-- ไม่พบ/เป็นของคนอื่น → o_category_id เป็น NULL
-- +goose StatementBegin
CREATE PROCEDURE public.oktp_category_update(
  IN p_category_id INTEGER, IN p_user_profile_id INTEGER,
  IN p_category_name TEXT, IN p_color TEXT, IN p_icon TEXT, IN p_category_type TEXT,
  OUT o_category_id INTEGER, OUT o_category_name TEXT, OUT o_color TEXT, OUT o_icon TEXT, OUT o_category_type TEXT,
  OUT o_created_at TIMESTAMPTZ, OUT o_updated_at TIMESTAMPTZ)
LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  UPDATE public.okdt_categories SET
    category_name = COALESCE(p_category_name, category_name),
    color = COALESCE(p_color, color),
    icon = COALESCE(p_icon, icon),
    category_type = COALESCE(p_category_type, category_type),
    updated_at = now()
  WHERE category_id = p_category_id AND user_profile_id = p_user_profile_id
  RETURNING category_id, category_name, color, icon, category_type, created_at, updated_at
  INTO o_category_id, o_category_name, o_color, o_icon, o_category_type, o_created_at, o_updated_at;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE PROCEDURE public.oktp_category_delete(IN p_category_id INTEGER, IN p_user_profile_id INTEGER, OUT o_deleted BOOLEAN)
LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  DELETE FROM public.okdt_categories WHERE category_id = p_category_id AND user_profile_id = p_user_profile_id;
  o_deleted := FOUND;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION public.oktf_master_category_list()
RETURNS TABLE (category_id INTEGER, category_name TEXT, color TEXT, icon TEXT, category_type TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql STABLE SET search_path = pg_catalog, pg_temp AS $$
  SELECT m.category_id, m.category_name, m.color, m.icon, m.category_type, m.created_at, m.updated_at
  FROM public.okdt_master_categories m ORDER BY m.category_name
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE PROCEDURE public.oktp_master_category_insert(
  IN p_category_name TEXT, IN p_color TEXT, IN p_icon TEXT, IN p_category_type TEXT,
  OUT o_category_id INTEGER, OUT o_created_at TIMESTAMPTZ, OUT o_updated_at TIMESTAMPTZ)
LANGUAGE plpgsql SET search_path = pg_catalog, pg_temp AS $$
BEGIN
  INSERT INTO public.okdt_master_categories (category_name, color, icon, category_type)
  VALUES (p_category_name, p_color, p_icon, p_category_type)
  RETURNING category_id, created_at, updated_at INTO o_category_id, o_created_at, o_updated_at;
END $$;
-- +goose StatementEnd

-- +goose Down
DROP PROCEDURE IF EXISTS public.oktp_master_category_insert(TEXT, TEXT, TEXT, TEXT);
DROP FUNCTION IF EXISTS public.oktf_master_category_list();
DROP PROCEDURE IF EXISTS public.oktp_category_delete(INTEGER, INTEGER);
DROP PROCEDURE IF EXISTS public.oktp_category_update(INTEGER, INTEGER, TEXT, TEXT, TEXT, TEXT);
DROP PROCEDURE IF EXISTS public.oktp_category_insert(TEXT, INTEGER, TEXT, TEXT, TEXT);
DROP PROCEDURE IF EXISTS public.oktp_user_insert(TEXT, TEXT, TEXT, TEXT);
DROP FUNCTION IF EXISTS public.oktf_user_credential_get(TEXT);
DROP FUNCTION IF EXISTS public.oktf_user_get(INTEGER);
