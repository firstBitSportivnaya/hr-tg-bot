-- Добавляем роли
INSERT INTO roles (role_name)
VALUES
    ('user'),
    ('manager'),
    ('admin');


-- Добавляем права
INSERT INTO permissions (permission_name)
VALUES
    ('start_test'),
    ('assign_test'),
    ('assign_hr'),
    ('assign_admin');


-- Получаем роли по имени
WITH role_user AS (SELECT id FROM roles WHERE role_name = 'user'),
     role_manager AS (SELECT id FROM roles WHERE role_name = 'manager'),
     role_admin AS (SELECT id FROM roles WHERE role_name = 'admin'),

-- Получаем права по имени
     perm_start_test AS (SELECT id FROM permissions WHERE permission_name = 'start_test'),
     perm_assign_test AS (SELECT id FROM permissions WHERE permission_name = 'assign_test'),
     perm_assign_hr AS (SELECT id FROM permissions WHERE permission_name = 'assign_hr'),
     perm_assign_admin AS (SELECT id FROM permissions WHERE permission_name = 'assign_admin')

-- Привязываем роли и права
INSERT INTO role_permissions (role_id, permission_id)
SELECT role_user.id, perm_start_test.id FROM role_user, perm_start_test
UNION ALL
SELECT role_manager.id, perm_start_test.id FROM role_manager, perm_start_test
UNION ALL
SELECT role_manager.id, perm_assign_test.id FROM role_manager, perm_assign_test
UNION ALL
SELECT role_admin.id, perm_start_test.id FROM role_admin, perm_start_test
UNION ALL
SELECT role_admin.id, perm_assign_test.id FROM role_admin, perm_assign_test
UNION ALL
SELECT role_admin.id, perm_assign_hr.id FROM role_admin, perm_assign_hr
UNION ALL
SELECT role_admin.id, perm_assign_admin.id FROM role_admin, perm_assign_admin;
