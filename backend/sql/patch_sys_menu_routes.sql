-- 补全「系统管理」下菜单（旧库升级时使用）
-- mysql --default-character-set=utf8mb4 -uroot -p jarvis < patch_sys_menu_routes.sql

USE jarvis;
SET NAMES utf8mb4;

UPDATE sys_menu SET status = '1', hidden = 0, is_deleted = 0 WHERE id IN (1, 2, 3, 6, 7);

UPDATE sys_menu SET
  name = '菜单管理', title = '菜单管理',
  route_path = '/system/menu/index', component_path = 'system/menu/index',
  status = '1', hidden = 0, is_deleted = 0, updated_time = NOW()
WHERE id = 6;

UPDATE sys_menu SET
  name = '字典管理', title = '字典管理',
  route_path = '/system/dict/index', component_path = 'system/dict/index',
  status = '1', hidden = 0, is_deleted = 0, updated_time = NOW()
WHERE id = 7;

INSERT IGNORE INTO sys_role_menus (role_id, menu_id) VALUES
  (1, 1), (1, 2), (1, 3), (1, 6), (1, 7);
