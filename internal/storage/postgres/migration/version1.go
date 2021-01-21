package migration

var version1 = `CREATE TYPE room_type AS ENUM ('private','public','broadcast');

CREATE TABLE IF NOT EXISTS "misc" (
	key VARCHAR (50) PRIMARY KEY,
	value VARCHAR (50) NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL
);

CREATE TABLE IF NOT EXISTS "room" (
	room_key VARCHAR (50) PRIMARY KEY,
	type room_type NOT NULL,
	created_by VARCHAR (50) NOT NULL,
	created_at timestamptz NULL
);

CREATE TABLE IF NOT EXISTS "user" (
	email VARCHAR (50) PRIMARY KEY,
	username VARCHAR(50) NOT NULL,
	name VARCHAR (50) NOT NULL,
	photo_url VARCHAR (150) NOT NULL
);

CREATE TABLE IF NOT EXISTS "user_room" (
	uuid VARCHAR (50) PRIMARY KEY,
	user_email VARCHAR (50) NOT NULL,
	room_key VARCHAR (50) NOT NULL
);

INSERT INTO public.misc ("key",value,created_at,updated_at) VALUES
('test','300',NULL,NULL)
;`
