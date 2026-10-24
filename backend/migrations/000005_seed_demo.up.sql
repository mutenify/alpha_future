-- 000005: демо-учётки и их компании для кнопок входа (DEMO_AUTH=true, пароль у всех: demo).
-- ИНН и реквизиты совпадают с ml/app/companies.json, поэтому мок отдаёт по ним те же данные.
-- Сценарии в scenario_hint проверены на текущих формулах скоринга и могут сдвинуться после правок математика.
-- Сгенерировано из companies.json; при изменении персон перегенерируйте файл, а не правьте руками.

INSERT INTO users (login, password_hash, role, full_name, position, scenario_hint) VALUES
    ('client1', crypt('demo', gen_salt('bf')), 'CLIENT', 'Иванова Ирина Сергеевна', 'Руководитель', 'Сценарий: 300 000 ₽ на 12 мес. — автоодобрение; 3 млн ₽ — контрпредложение'),
    ('client2', crypt('demo', gen_salt('bf')), 'CLIENT', 'Смирнов Павел Андреевич', 'Руководитель', 'Сценарий: 3 млн ₽ на 12 мес. — контрпредложение, варианты А и Б'),
    ('client3', crypt('demo', gen_salt('bf')), 'CLIENT', 'Кузнецова Мария Олеговна', 'Руководитель', 'Сценарий: ручной разбор инспектором (PD 10–25%)'),
    ('client4', crypt('demo', gen_salt('bf')), 'CLIENT', 'Попов Алексей Викторович', 'Руководитель', 'Сценарий: отказ, высокий PD'),
    ('client5', crypt('demo', gen_salt('bf')), 'CLIENT', 'Орлов Дмитрий Николаевич', 'Руководитель', 'Сценарий: отказ по стоп-фактору (ликвидация)'),
    ('client6', crypt('demo', gen_salt('bf')), 'CLIENT', 'Васильев Сергей Ильич', 'Руководитель', 'Сценарий: ФНС недоступна — ручной разбор'),
    ('client7', crypt('demo', gen_salt('bf')), 'CLIENT', 'Хабибуллин Ринат Маратович', 'Руководитель', 'Сценарий: эталонная заявка модели v1'),
    ('client8', crypt('demo', gen_salt('bf')), 'CLIENT', 'Мельникова Елена Игоревна', 'Руководитель', 'Сценарий: эталонная заявка модели v1'),
    ('client9', crypt('demo', gen_salt('bf')), 'CLIENT', 'Бондаренко Артём Сергеевич', 'Руководитель', 'Сценарий: эталонная заявка модели v1'),
    ('client10', crypt('demo', gen_salt('bf')), 'CLIENT', 'Соколова Анна Викторовна', 'Руководитель', 'Сценарий: эталонная заявка модели v1'),
    ('inspector1', crypt('demo', gen_salt('bf')), 'INSPECTOR', 'Елена Волкова', 'Андеррайтер МСБ', 'Сценарий: очередь и ручной разбор')
ON CONFLICT (login) DO UPDATE SET
    full_name = EXCLUDED.full_name,
    position = EXCLUDED.position,
    scenario_hint = EXCLUDED.scenario_hint;

INSERT INTO companies (owner_user_id, inn, ogrn, name, legal_form, okved, registered_at, settlement_account)
SELECT u.id, s.inn, s.ogrn, s.name, s.legal_form, s.okved, s.registered_at, s.settlement_account
FROM (VALUES
    ('client1', '7701234560', '1027700132195', 'ООО «Колос и Ко»', 'LLC', '10.71', DATE '2022-03-14', '40702810400000012345'),
    ('client2', '7812345675', '1037828001234', 'ООО «Северная Линия»', 'LLC', '49.41', DATE '2018-07-02', '40702810500000022346'),
    ('client3', '5401234566', '1165476123456', 'ООО «Сибирский Фермер»', 'LLC', '01.11', DATE '2016-05-19', '40702810600000032347'),
    ('client4', '6671234569', '1086671012345', 'ООО «Урал Деталь»', 'LLC', '25.62', DATE '2008-11-06', '40702810700000042348'),
    ('client5', '5251234564', '1125257004321', 'ООО «Ока Мебель»', 'LLC', '31.09', DATE '2012-02-28', '40702810100000072341'),
    ('client6', '7700000503', '1843539811688', 'ООО «Маяк»', 'LLC', '70.22', DATE '2012-01-10', '40702810406723386359'),
    ('client7', '1651234561', '1141690067890', 'ООО «Волга Текстиль»', 'LLC', '13.92', DATE '2014-09-23', '40702810800000052349'),
    ('client8', '2311234566', '1052305701234', 'ООО «Южный Сад»', 'LLC', '46.31', DATE '2005-04-12', '40702810900000062340'),
    ('client9', '6161234561', '1196196012345', 'ООО «Дон Строй»', 'LLC', '41.20', DATE '2019-08-15', '40702810200000082342'),
    ('client10', '7707083893', '1027700123456', 'ООО «Цифровая Среда»', 'LLC', '62.01', DATE '2017-06-20', '40702810300000092343')
) AS s (login, inn, ogrn, name, legal_form, okved, registered_at, settlement_account)
JOIN users u ON u.login = s.login
ON CONFLICT (inn) DO UPDATE SET
    owner_user_id = EXCLUDED.owner_user_id,
    name = EXCLUDED.name,
    okved = EXCLUDED.okved,
    settlement_account = EXCLUDED.settlement_account;

-- Готовая очередь инспектора из макетов. Фиксированные UUID позволяют безопасно
-- повторять seed, а OVERRIDING SYSTEM VALUE сохраняет отображаемые Z-152…Z-156.
INSERT INTO applications
    (id, number, client_user_id, company_id, product_id, amount, term_months,
     purpose, status, assigned_inspector_id, submitted_at, created_at, updated_at)
OVERRIDING SYSTEM VALUE
SELECT s.id::uuid, s.number, client.id, company.id, product.id, s.amount,
       s.term_months, s.purpose, s.status, inspector.id, s.created_at,
       s.created_at, s.created_at
FROM (VALUES
    ('00000000-0000-0000-0000-000000000152', 152::bigint, 'client3', 'INVESTMENT_LOAN', 5000000.00, 36, 'Покупка оборудования', 'MANUAL_REVIEW',  TIMESTAMPTZ '2026-10-20 09:15:00+03'),
    ('00000000-0000-0000-0000-000000000153', 153::bigint, 'client6', 'WORKING_CAPITAL', 1000000.00, 12, 'Пополнение оборотных средств', 'INFO_REQUESTED', TIMESTAMPTZ '2026-10-20 10:30:00+03'),
    ('00000000-0000-0000-0000-000000000154', 154::bigint, 'client2', 'WORKING_CAPITAL', 3000000.00, 12, 'Закупка сырья', 'COUNTER_OFFER',  TIMESTAMPTZ '2026-10-20 11:45:00+03'),
    ('00000000-0000-0000-0000-000000000155', 155::bigint, 'client1', 'WORKING_CAPITAL',  300000.00, 12, 'Пополнение оборотных средств', 'APPROVED',       TIMESTAMPTZ '2026-10-20 13:00:00+03'),
    ('00000000-0000-0000-0000-000000000156', 156::bigint, 'client4', 'WORKING_CAPITAL', 8000000.00, 24, 'Расширение производства', 'REJECTED',       TIMESTAMPTZ '2026-10-20 14:20:00+03')
) AS s (id, number, login, product_code, amount, term_months, purpose, status, created_at)
JOIN users client ON client.login = s.login
JOIN companies company ON company.owner_user_id = client.id
JOIN credit_products product ON product.code = s.product_code
CROSS JOIN users inspector
WHERE inspector.login = 'inspector1'
ON CONFLICT (id) DO NOTHING;

SELECT setval(
    pg_get_serial_sequence('applications', 'number'),
    GREATEST((SELECT max(number) FROM applications), 156),
    true
);

INSERT INTO application_status_history
    (application_id, from_status, to_status, actor_type, reason, created_at)
SELECT a.id, NULL, a.status, 'SYSTEM', 'Демонстрационная заявка', a.created_at
FROM applications a
WHERE a.id IN (
    '00000000-0000-0000-0000-000000000152',
    '00000000-0000-0000-0000-000000000153',
    '00000000-0000-0000-0000-000000000154',
    '00000000-0000-0000-0000-000000000155',
    '00000000-0000-0000-0000-000000000156'
)
AND NOT EXISTS (
    SELECT 1 FROM application_status_history h
    WHERE h.application_id = a.id AND h.reason = 'Демонстрационная заявка'
);
