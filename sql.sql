-- DROP TABLE IF EXISTS users;
CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY,
    name VARCHAR(32),
    days_without_weed INT,
    karma INT,
    subscriptions text []
);

-- 

INSERT INTO users (id, name, days_without_weed, karma, subscriptions)
VALUES (
    282857241,
    'humanityForBegginers',
    0,
    0,
    '{1,2,3}'
);

INSERT INTO users (id, name, days_without_weed, karma, subscriptions)
VALUES (
    90217964,
    'nbox363',
    0,
    0,
    '{1,2,3}'
);