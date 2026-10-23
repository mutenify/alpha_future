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
    ('client6', '7700000503', '1843539811688', 'ООО «Маяк»', 'LLC', '70.22', DATE '2012-01-10', '40702810406723386359')
) AS s (login, inn, ogrn, name, legal_form, okved, registered_at, settlement_account)
JOIN users u ON u.login = s.login
ON CONFLICT (inn) DO UPDATE SET
    owner_user_id = EXCLUDED.owner_user_id,
    name = EXCLUDED.name,
    okved = EXCLUDED.okved,
    settlement_account = EXCLUDED.settlement_account;
