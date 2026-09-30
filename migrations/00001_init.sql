-- ตารางที่ repository ตัวอย่าง (users, auth, categories) ใช้
-- +goose Up
CREATE TABLE okdt_user_profiles (
  user_profile_id SERIAL PRIMARY KEY,
  email           TEXT NOT NULL UNIQUE,
  full_name       TEXT NOT NULL,
  password_hash   TEXT NOT NULL,
  role            TEXT NOT NULL DEFAULT 'user',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE okdt_master_categories (
  category_id   SERIAL PRIMARY KEY,
  category_name TEXT NOT NULL,
  color         TEXT NOT NULL DEFAULT '',
  icon          TEXT NOT NULL DEFAULT '',
  category_type TEXT NOT NULL CHECK (category_type IN ('income', 'expense', 'both')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE okdt_categories (
  category_id     SERIAL PRIMARY KEY,
  category_name   TEXT NOT NULL,
  user_profile_id INTEGER NOT NULL REFERENCES okdt_user_profiles (user_profile_id) ON DELETE CASCADE,
  color           TEXT NOT NULL DEFAULT '',
  icon            TEXT NOT NULL DEFAULT '',
  category_type   TEXT NOT NULL CHECK (category_type IN ('income', 'expense', 'both')),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX okdt_categories_user_idx ON okdt_categories (user_profile_id);

-- +goose StatementBegin
CREATE FUNCTION public.oktf_category_get(p_user_profile_id INTEGER)
RETURNS SETOF public.okdt_categories
LANGUAGE sql STABLE SET search_path = pg_catalog, pg_temp AS $$
  SELECT * FROM public.okdt_categories WHERE user_profile_id = p_user_profile_id ORDER BY category_name, category_id
$$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS oktf_category_get(INTEGER);
DROP TABLE IF EXISTS okdt_categories;
DROP TABLE IF EXISTS okdt_master_categories;
DROP TABLE IF EXISTS okdt_user_profiles;
