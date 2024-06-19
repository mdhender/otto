--  Copyright (c) 2024 Michael D Henderson. All rights reserved.

INSERT INTO migrations (id)
values ('20240619131300_initialization');

INSERT INTO paths (name, path)
VALUES ('assets', 'assets');
INSERT INTO paths (name, path)
VALUES ('templates', 'templates');

INSERT INTO users (handle, hashed_password, clan, magic, enabled)
VALUES ('otto', '', '0000', '', 'Y');

