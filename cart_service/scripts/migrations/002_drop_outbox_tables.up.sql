-- Удаление таблицы outbox (если она существует)
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS outbox;

-- Удаление последовательностей, если они есть
DROP SEQUENCE IF EXISTS outbox_events_id_seq;
DROP SEQUENCE IF EXISTS outbox_id_seq; 