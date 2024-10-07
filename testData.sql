INSERT INTO cleni (jmeno, discord_id) VALUES ('Test', '1234567890');
INSERT INTO cleni (jmeno, discord_id) VALUES ('Test2', '0987654321');
INSERT INTO schuzky (nazev, kdy, upozorneno) VALUES ('Test', 1729440000, 0);
INSERT INTO schuzky (nazev, kdy, upozorneno) VALUES ('Test2', 1730048400, 0);
INSERT INTO porada (cleni_id, schuzky_id, zprava_id) VALUES (1, 1, '');
INSERT INTO porada (cleni_id, schuzky_id, zprava_id) VALUES (2, 2, '');
INSERT INTO porada (cleni_id, schuzky_id, zprava_id) VALUES (1, 2, '');