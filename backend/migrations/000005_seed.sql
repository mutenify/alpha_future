-- Demo data for the end-to-end decision and inspector-queue scenarios.
-- Requires pgcrypto and the users/companies/applications schema from migrations 000001-000004.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (email, password_hash, full_name, role, scenario_hint)
VALUES
    ('auto@demo.local', crypt('demo', gen_salt('bf')), 'Алексей Автоодобрение', 'CLIENT', 'AUTO_APPROVE'),
    ('counter-a@demo.local', crypt('demo', gen_salt('bf')), 'Анна Контрпредложение А', 'CLIENT', 'COUNTER_OFFER_A'),
    ('counter-b@demo.local', crypt('demo', gen_salt('bf')), 'Борис Контрпредложение Б', 'CLIENT', 'COUNTER_OFFER_B'),
    ('manual@demo.local', crypt('demo', gen_salt('bf')), 'Мария Ручная проверка', 'CLIENT', 'MANUAL_REVIEW'),
    ('documents@demo.local', crypt('demo', gen_salt('bf')), 'Ирина Запрос документов', 'CLIENT', 'INFO_REQUESTED'),
    ('liquidation@demo.local', crypt('demo', gen_salt('bf')), 'Леонид Ликвидация', 'CLIENT', 'LIQUIDATION_STOP'),
    ('source-down@demo.local', crypt('demo', gen_salt('bf')), 'Сергей Недоступный источник', 'CLIENT', 'SOURCE_UNAVAILABLE'),
    ('inspector@demo.local', crypt('demo', gen_salt('bf')), 'Елена Волкова', 'INSPECTOR', 'MANUAL_REVIEW')
ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name,
    role = EXCLUDED.role,
    scenario_hint = EXCLUDED.scenario_hint;

INSERT INTO companies (owner_id, inn, name, legal_form)
SELECT u.id, seed.inn, seed.name, seed.legal_form
FROM (VALUES
    ('auto@demo.local',        '7707083893',   'ООО «Альфа Доставка»',       'OOO'),
    ('counter-a@demo.local',   '7710140679',   'ООО «Точка Роста»',          'OOO'),
    ('counter-b@demo.local',   '500100732259', 'ИП Смирнова Виктория Олеговна', 'IP'),
    ('manual@demo.local',      '7736050003',   'ООО «Северный Ветер»',       'OOO'),
    ('documents@demo.local',   '7811005549',   'ООО «Невская Линия»',        'OOO'),
    ('liquidation@demo.local', '6671000017',   'ООО «Урал Комплект»',        'OOO'),
    ('source-down@demo.local', '503606511370', 'ИП Кузнецов Артём Ильич',    'IP')
) AS seed(email, inn, name, legal_form)
JOIN users u ON u.email = seed.email
ON CONFLICT (inn) DO UPDATE SET
    owner_id = EXCLUDED.owner_id,
    name = EXCLUDED.name,
    legal_form = EXCLUDED.legal_form;

INSERT INTO applications
    (number, company_id, status, requested_amount, assigned_inspector_id)
SELECT
    seed.number,
    c.id,
    seed.status,
    seed.requested_amount,
    CASE WHEN seed.number >= 'Z-152' THEN inspector.id END
FROM (VALUES
    ('Z-150', '7707083893',   'APPROVED',        3000000.00),
    ('Z-151', '7710140679',   'COUNTER_OFFERED', 5000000.00),
    ('Z-152', '500100732259', 'COUNTER_OFFERED', 7500000.00),
    ('Z-153', '7736050003',   'MANUAL_REVIEW',   4200000.00),
    ('Z-154', '7811005549',   'INFO_REQUESTED',  2800000.00),
    ('Z-155', '6671000017',   'REJECTED',        6100000.00),
    ('Z-156', '503606511370', 'MANUAL_REVIEW',   3500000.00)
) AS seed(number, inn, status, requested_amount)
JOIN companies c ON c.inn = seed.inn
CROSS JOIN (SELECT id FROM users WHERE email = 'inspector@demo.local') AS inspector
ON CONFLICT (number) DO UPDATE SET
    company_id = EXCLUDED.company_id,
    status = EXCLUDED.status,
    requested_amount = EXCLUDED.requested_amount,
    assigned_inspector_id = EXCLUDED.assigned_inspector_id;
